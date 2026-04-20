local jwt = require "resty.jwt"
local cjson = require "cjson.safe"
local redis = require "resty.redis"

local secret = os.getenv("JWT_SECRET") or "dev-secret"
local auth = ngx.var.http_authorization
if not auth or not auth:find("Bearer ") then
  ngx.status = 401
  ngx.say(cjson.encode({error="missing bearer token"}))
  return ngx.exit(401)
end

local token = auth:gsub("Bearer%s+", "")
local jwt_obj = jwt:verify(secret, token)
if not jwt_obj.verified then
  ngx.status = 401
  ngx.say(cjson.encode({error="invalid token"}))
  return ngx.exit(401)
end

local payload = jwt_obj.payload or {}
local exp = tonumber(payload.exp or "0")
if exp <= ngx.time() then
  ngx.status = 401
  ngx.say(cjson.encode({error="token expired"}))
  return ngx.exit(401)
end

local red = redis:new()
red:set_timeout(100)
local ok, err = red:connect(os.getenv("VALKEY_HOST") or "valkey", tonumber(os.getenv("VALKEY_PORT") or "6379"))
if not ok then
  ngx.log(ngx.ERR, "redis connect failed: ", err)
else
  local key = "jwt:blacklist:" .. token
  local banned = red:get(key)
  if banned and banned ~= ngx.null then
    ngx.status = 401
    ngx.say(cjson.encode({error="token blacklisted"}))
    return ngx.exit(401)
  end
end
ngx.req.set_header("X-User-Id", payload.sub or payload.userId or "unknown")
