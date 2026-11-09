package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/PavelLymin/diploma/services/auth/internal/domain"
	"github.com/google/uuid"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func New(client *redis.Client) *RedisRepository {
	if client == nil {
		panic("redis client is required")
	}

	return &RedisRepository{
		client: client,
	}
}

type challengeRecord struct {
	ID                string `json:"id"`
	Phone             string `json:"phone"`
	CodeHash          string `json:"code_hash"`
	Attempts          int    `json:"attempts"`
	CreatedAtMS       int64  `json:"created_at_ms"`
	ExpiresAtMS       int64  `json:"expires_at_ms"`
	ResendAvailableMS int64  `json:"resend_available_at_ms"`
	BlockedUntilMS    int64  `json:"blocked_until_ms,omitempty"`
}

type idempotencyRecord struct {
	Fingerprint string           `json:"fingerprint"`
	Challenge   *challengeRecord `json:"challenge"`
}

var createChallengeScript = redis.NewScript(`
local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)

local idemRaw = redis.call('GET', KEYS[4])
if idemRaw then
	local idem = cjson.decode(idemRaw)
	if idem.fingerprint == ARGV[5] then
		return {"2", cjson.encode(idem.challenge)}
	else
		return {"-2", ""}
	end
end

local blockedUntil = redis.call('GET', KEYS[3])
if blockedUntil then
	return {"-1", blockedUntil}
end

local activeID = redis.call('GET', KEYS[2])
if activeID then
	local activeKey = KEYS[5] .. activeID
	local activeRaw = redis.call('GET', activeKey)
	if activeRaw then
		local active = cjson.decode(activeRaw)
		if tonumber(active.expires_at_ms) > now then
            local resendAt = tonumber(active.resend_available_at_ms)
			if resendAt > now then
                return {"-3", tostring(resendAt)}
            end
            return {"-4", tostring(resendAt)}
		end
		redis.call('DEL', activeKey)
	end
	redis.call('DEL', KEYS[2])
end

local challenge = cjson.decode(ARGV[1])
challenge.created_at_ms = now
challenge.expires_at_ms = now + tonumber(ARGV[4])
challenge.resend_available_at_ms = now + tonumber(ARGV[7])
local challengeJSON = cjson.encode(challenge)
local idem = cjson.encode({fingerprint=ARGV[5], challenge=challenge})
redis.call('SET', KEYS[1], challengeJSON, 'PX',tonumber(ARGV[4]))
redis.call('SET', KEYS[2], ARGV[2], 'PX', tonumber(ARGV[4]))
redis.call('SET', KEYS[4], idem, 'PX',tonumber(ARGV[6]))
return {"1", challengeJSON}
`)

var resendChallengeScript = redis.NewScript(`
local t = redis.call("TIME")
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)

local raw = redis.call("GET", KEYS[4])
if raw then
	local idem = cjson.decode(raw)
	if idem.fingerprint ~= ARGV[5] then
		return {"-2", ""}
	else
		return {"2", cjson.encode(idem.challenge)}
	end
end

local blockedUntil = redis.call('GET', KEYS[3])
if blockedUntil then
	return {"-1", blockedUntil}
end

local oldRaw = redis.call("GET", KEYS[6])
if not oldRaw then
	return {"-5", ""}
end
local old = cjson.decode(oldRaw)
if old.phone ~= ARGV[3] then
	return {"-8", ""}
end
if tonumber(old.expires_at_ms) <= now then
	return {"-7", ""}
end
local activeID = redis.call('GET', KEYS[2])
if activeID ~= ARGV[7] then
	return {"-6", ""}
end
local resendAt = tonumber(old.resend_available_at_ms)
if resendAt > now then
	return {"-3", tostring(resendAt)}
end

local challenge = cjson.decode(ARGV[1])
challenge.created_at_ms = now
challenge.expires_at_ms = now + tonumber(ARGV[4])
challenge.resend_available_at_ms = now + tonumber(ARGV[8])
local challengeJSON = cjson.encode(challenge)
local idem = cjson.encode({fingerprint=ARGV[5], challenge=challenge})
redis.call('SET', KEYS[1], challengeJSON, 'PX',tonumber(ARGV[4]))
redis.call('DEL', KEYS[6])
redis.call('SET', KEYS[2], ARGV[2], 'PX',tonumber(ARGV[4]))
redis.call('SET', KEYS[4], idem, 'PX',tonumber(ARGV[6]))
return {"1", challengeJSON}
`)

var verifyChallengeScript = redis.NewScript(`
local function compare_time_equal(a, b)
	if type(a) ~= "string" or type(b) ~= "string" then
		return false
	end

	local len = #a
	if len ~= #b then
		return false
	end

	local diff = 0

	for i = 1, len do
		local delta = bit.bxor(
            string.byte(a, i),
            string.byte(b, i)
        )
		diff = diff + delta * delta
	end

	return diff == 0
end

local raw = redis.call("GET", KEYS[1])
if not raw then
	return {"-5", ""}
end
local challenge = cjson.decode(raw)

local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)

local phoneBlock = redis.call('GET', KEYS[3])
if phoneBlock then
	return {"-2", phoneBlock}
end
if tonumber(challenge.expires_at_ms) <= now then
	redis.call('DEL', KEYS[1])
	if redis.call('GET', KEYS[2]) == challenge.id then
		redis.call('DEL', KEYS[2])
	end
	return {"-3", ""}
end
if tonumber(challenge.blocked_until_ms or 0) > now then
	return {"-2", tostring(challenge.blocked_until_ms)}
end
if tonumber(challenge.attempts) >= tonumber(ARGV[2]) then
	return {"-4", ""}
end

if not compare_time_equal(ARGV[1], challenge.code_hash) then
	challenge.attempts = tonumber(challenge.attempts) + 1
	if challenge.attempts >= tonumber(ARGV[2]) then
		local blockUntil = now + tonumber(ARGV[3])
		challenge.blocked_until_ms = blockUntil
		redis.call('SET', KEYS[3], tostring(blockUntil), "PX", tonumber(ARGV[3]))
		redis.call('SET', KEYS[1], cjson.encode(challenge), "KEEPTTL")
		return {"2", tostring(blockUntil)}
	end
	redis.call('SET', KEYS[1], cjson.encode(challenge), "KEEPTTL")
	return {"1", tostring(challenge.attempts)}
end

redis.call('DEL', KEYS[1])
if redis.call('GET', KEYS[2]) == challenge.id then
	redis.call('DEL', KEYS[2])
end
return {"0", raw}
`)

var deleteChallengeScript = redis.NewScript(`
local raw = redis.call("GET", KEYS[3])
if not raw then return 0 end
local idem = cjson.decode(raw)
if not idem.challenge or idem.challenge.id ~= ARGV[1] then return 0 end
if redis.call("GET", KEYS[2]) == ARGV[1] then
	redis.call("DEL", KEYS[2])
end
redis.call("DEL", KEYS[1])
redis.call("DEL", KEYS[3])
return 1
`)

func (r *RedisRepository) CreateChallenge(
	ctx context.Context,
	requestID uuid.UUID,
	fingerprint string,
	challenge *domain.OTPChallenge,
	challengeTTL, idempotencyTTL time.Duration,
) (*domain.OTPChallenge, bool, error) {
	record := toRecord(challenge)
	data, err := json.Marshal(record)
	if err != nil {
		return nil, false, fmt.Errorf("marshal OTP challenge: %w", err)
	}

	result := createChallengeScript.Run(ctx, r.client, []string{
		r.challengeKey(challenge.ID),
		r.phoneKey(challenge.Phone),
		r.phoneBlockKey(challenge.Phone),
		r.idempotencyKey("create", requestID),
		r.challengePrefix(),
	},
		string(data),
		challenge.ID.String(),
		challenge.Phone,
		challengeTTL.Milliseconds(),
		fingerprint,
		idempotencyTTL.Milliseconds(),
		challenge.ResendAvailableAt.Sub(challenge.CreatedAt).Milliseconds(),
	)
	status, payload, err := parseTuple(result)
	if err != nil {
		return nil, false, err
	}
	switch status {
	case 1, 2:
		created, err := decodeChallenge([]byte(payload))
		return created, status == 1, err
	case -1:
		return nil, false, r.rateLimitError(ctx, domain.RateLimitCodePhoneBlocked, payload, true)
	case -2:
		return nil, false, domain.ErrIdempotencyConflict
	case -3:
		return nil, false, r.rateLimitError(ctx, domain.RateLimitCodeResendTooEarly, payload, false)
	case -4:
		return nil, false, domain.ErrChallengeAlreadyActive
	default:
		return nil, false, fmt.Errorf("unexpected create challenge result: %d", status)
	}
}

func (r *RedisRepository) ResendChallenge(
	ctx context.Context,
	requestID, oldChallengeID uuid.UUID,
	fingerprint string,
	challenge *domain.OTPChallenge,
	challengeTTL, idempotencyTTL time.Duration,
) (*domain.OTPChallenge, bool, error) {
	if oldChallengeID == uuid.Nil {
		return nil, false, domain.ErrChallengeNotFound
	}
	record := toRecord(challenge)
	data, err := json.Marshal(record)
	if err != nil {
		return nil, false, fmt.Errorf("marshal replacement OTP challenge: %w", err)
	}
	result := resendChallengeScript.Run(ctx, r.client, []string{
		r.challengeKey(challenge.ID),
		r.phoneKey(challenge.Phone),
		r.phoneBlockKey(challenge.Phone),
		r.idempotencyKey("resend", requestID),
		r.challengePrefix(),
		r.challengeKey(oldChallengeID),
	},
		string(data),
		challenge.ID.String(),
		challenge.Phone,
		challengeTTL.Milliseconds(),
		fingerprint,
		idempotencyTTL.Milliseconds(),
		oldChallengeID.String(),
		challenge.ResendAvailableAt.Sub(challenge.CreatedAt).Milliseconds(),
	)
	status, payload, err := parseTuple(result)
	if err != nil {
		return nil, false, err
	}
	switch status {
	case 1, 2:
		created, err := decodeChallenge([]byte(payload))
		return created, status == 1, err
	case -1:
		return nil, false, r.rateLimitError(ctx, domain.RateLimitCodePhoneBlocked, payload, true)
	case -2:
		return nil, false, domain.ErrIdempotencyConflict
	case -3:
		return nil, false, r.rateLimitError(ctx, domain.RateLimitCodeResendTooEarly, payload, false)
	case -5:
		return nil, false, domain.ErrChallengeNotFound
	case -6:
		return nil, false, domain.ErrStaleChallenge
	case -7:
		return nil, false, domain.ErrChallengeExpired
	case -8:
		return nil, false, errors.New("challenge phone mismatch")
	default:
		return nil, false, fmt.Errorf("unexpected resend challenge result: %d", status)
	}
}

func (r *RedisRepository) VerifyAndConsume(
	ctx context.Context,
	challengeID uuid.UUID,
	codeHash string,
	maxAttempts int,
	blockDuration time.Duration,
) (*domain.OTPChallenge, error) {
	current, err := r.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	result := verifyChallengeScript.Run(ctx, r.client, []string{
		r.challengeKey(challengeID),
		r.phoneKey(current.Phone),
		r.phoneBlockKey(current.Phone),
	},
		codeHash,
		maxAttempts,
		blockDuration.Milliseconds(),
	)
	status, payload, err := parseTuple(result)
	if err != nil {
		return nil, err
	}
	switch status {
	case 0:
		return decodeChallenge([]byte(payload))
	case 1:
		return nil, domain.ErrInvalidOTP
	case 2:
		return nil, r.rateLimitError(ctx, domain.RateLimitCodePhoneBlocked, payload, true)
	case -2:
		return nil, r.rateLimitError(ctx, domain.RateLimitCodePhoneBlocked, payload, true)
	case -3:
		return nil, domain.ErrChallengeExpired
	case -4:
		return nil, domain.ErrMaxAttemptsExceeded
	case -5:
		return nil, domain.ErrChallengeNotFound
	default:
		return nil, fmt.Errorf("unexpected verify challenge result: %d", status)
	}
}

func (r *RedisRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.OTPChallenge, error) {
	data, err := r.client.Get(ctx, r.challengeKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrChallengeNotFound
		}
		return nil, fmt.Errorf("get OTP challenge from Redis: %w", err)
	}
	return decodeChallenge(data)
}

func (r *RedisRepository) GetIdempotentChallenge(
	ctx context.Context,
	operation string,
	requestID uuid.UUID,
	fingerprint string,
) (*domain.OTPChallenge, bool, error) {
	data, err := r.client.Get(ctx, r.idempotencyKey(operation, requestID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get OTP idempotency record: %w", err)
	}

	var record idempotencyRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, false, fmt.Errorf("decode OTP idempotency record: %w", err)
	}
	if record.Fingerprint != fingerprint {
		return nil, false, domain.ErrIdempotencyConflict
	}
	if record.Challenge == nil {
		return nil, false, errors.New("OTP idempotency record has no challenge")
	}
	challengeJSON, err := json.Marshal(record.Challenge)
	if err != nil {
		return nil, false, fmt.Errorf("marshal idempotent OTP challenge: %w", err)
	}
	challenge, err := decodeChallenge(challengeJSON)
	if err != nil {
		return nil, false, err
	}
	return challenge, true, nil
}

func (r *RedisRepository) IsPhoneBlocked(
	ctx context.Context,
	phone string,
) (*time.Time, error) {
	value, err := r.client.Get(ctx, r.phoneBlockKey(phone)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get phone block: %w", err)
	}

	blockedUntil, err := r.parseTimestamp(value)
	if err != nil {
		return nil, err
	}

	return &blockedUntil, nil
}

func (r *RedisRepository) DeletePending(
	ctx context.Context,
	operation string,
	requestID,
	challengeID uuid.UUID,
	phone string,
) error {
	deleted, err := deleteChallengeScript.Run(ctx, r.client, []string{
		r.challengeKey(challengeID),
		r.phoneKey(phone),
		r.idempotencyKey(operation, requestID),
	},
		challengeID.String(),
	).Int()
	if err != nil {
		return fmt.Errorf("delete pending OTP challenge: %w", err)
	}
	if deleted != 1 {
		return fmt.Errorf("pending OTP challenge %s was not deleted", challengeID)
	}
	return nil
}

func (r *RedisRepository) ServerTime(ctx context.Context) (time.Time, error) {
	result, err := r.client.Time(ctx).Result()
	if err != nil {
		return time.Time{}, fmt.Errorf("get Redis server time: %w", err)
	}
	return result.UTC(), nil
}

func toRecord(c *domain.OTPChallenge) *challengeRecord {
	record := &challengeRecord{
		ID:                c.ID.String(),
		Phone:             c.Phone,
		CodeHash:          c.CodeHash,
		Attempts:          c.Attempts,
		CreatedAtMS:       c.CreatedAt.UTC().UnixMilli(),
		ExpiresAtMS:       c.ExpiresAt.UTC().UnixMilli(),
		ResendAvailableMS: c.ResendAvailableAt.UTC().UnixMilli(),
	}
	if c.BlockedUntil != nil {
		record.BlockedUntilMS = c.BlockedUntil.UTC().UnixMilli()
	}
	return record
}

func decodeChallenge(data []byte) (*domain.OTPChallenge, error) {
	var record challengeRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("decode OTP challenge: %w", err)
	}
	id, err := uuid.Parse(record.ID)
	if err != nil {
		return nil, fmt.Errorf("decode OTP challenge ID: %w", err)
	}
	c := &domain.OTPChallenge{
		ID:                id,
		Phone:             record.Phone,
		CodeHash:          record.CodeHash,
		Attempts:          record.Attempts,
		CreatedAt:         time.UnixMilli(record.CreatedAtMS).UTC(),
		ExpiresAt:         time.UnixMilli(record.ExpiresAtMS).UTC(),
		ResendAvailableAt: time.UnixMilli(record.ResendAvailableMS).UTC(),
	}
	if record.BlockedUntilMS > 0 {
		t := time.UnixMilli(record.BlockedUntilMS).UTC()
		c.BlockedUntil = &t
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf(
			"stored OTP challenge is corrupted: %v",
			err,
		)
	}
	return c, nil
}

func parseTuple(cmd *redis.Cmd) (int64, string, error) {
	values, err := cmd.StringSlice()
	if err != nil {
		return 0, "", fmt.Errorf("parse Redis response: %w", err)
	}

	if len(values) != 2 {
		return 0, "", fmt.Errorf("unexpected Redis response length: %d", len(values))
	}

	status, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("parse Redis response status: %w", err)
	}
	return status, values[1], nil
}

func (r *RedisRepository) rateLimitError(
	ctx context.Context,
	reason string,
	payload string,
	blocked bool,
) error {
	until, parseErr := r.parseTimestamp(payload)
	now, err := r.ServerTime(ctx)
	if err != nil {
		return fmt.Errorf("get server time for rate limit: %w", err)
	}

	if parseErr != nil {
		if blocked {
			return fmt.Errorf("parse blocked timestamp: %w", parseErr)
		} else {
			return fmt.Errorf("parse rate limit timestamp %q: %w", payload, parseErr)
		}
	}

	if blocked {
		return domain.NewRateLimitError(reason, now, nil, &until)
	}
	return domain.NewRateLimitError(reason, now, &until, nil)
}

func (r *RedisRepository) parseTimestamp(value string) (time.Time, error) {
	ms, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"invalid timestamp %q: %w",
			value,
			err,
		)
	}

	return time.UnixMilli(ms).UTC(), nil
}

func (r *RedisRepository) challengePrefix() string { return "auth:challenge:" }

// Данные challenge
func (r *RedisRepository) challengeKey(id uuid.UUID) string {
	return r.challengePrefix() + id.String()
}

// Активный challenge конкретного номера телефона
func (r *RedisRepository) phoneKey(phone string) string {
	return "auth:phone_to_challenge:" + phone
}

// Блокировка телефона
func (r *RedisRepository) phoneBlockKey(phone string) string {
	return "auth:phone_block:" + phone
}

// Результат идемпотентной операции
func (r *RedisRepository) idempotencyKey(operation string, id uuid.UUID) string {
	return "auth:otp_idempotency:" + operation + ":" + id.String()
}
