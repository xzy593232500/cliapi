import Fastify from 'fastify';
import cors from '@fastify/cors';
import { Pool } from 'pg';
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import crypto from 'crypto';

const app = Fastify({ logger: true });
await app.register(cors, { origin: true });

const pool = new Pool({ connectionString: process.env.DATABASE_URL });
const JWT_SECRET = process.env.JWT_SECRET || 'change-me';
const API_KEY_PREFIX = process.env.API_KEY_PREFIX || 'ocp';
const OPENAI_BASE_URL = (process.env.OPENAI_BASE_URL || 'https://api.openai.com/v1').replace(/\/$/, '');
const OPENAI_API_KEY = process.env.OPENAI_API_KEY || '';
const TOKEN_RATE_TOKENS = Number(process.env.TOKEN_RATE_TOKENS || 500000);
const TOKEN_RATE_CENTS = Number(process.env.TOKEN_RATE_CENTS || 100);
const ALLOW_REGISTRATION = (process.env.ALLOW_REGISTRATION || 'true') === 'true';
const PORT = Number(process.env.PORT || 8787);

function sha256(value) {
  return crypto.createHash('sha256').update(value).digest('hex');
}

function genId() {
  return crypto.randomUUID();
}

function signUser(user) {
  return jwt.sign({ sub: user.id, role: user.role, email: user.email }, JWT_SECRET, { expiresIn: '7d' });
}

function costCents(totalTokens) {
  return Math.ceil((Number(totalTokens || 0) * TOKEN_RATE_CENTS) / TOKEN_RATE_TOKENS);
}

function parseBearer(value = '') {
  const m = value.match(/^Bearer\s+(.+)$/i);
  return m ? m[1] : null;
}

async function authUser(req, reply) {
  const token = parseBearer(req.headers.authorization);
  if (!token) return reply.code(401).send({ error: 'missing_authorization' });
  try {
    const payload = jwt.verify(token, JWT_SECRET);
    const { rows } = await pool.query('select id, email, role, balance_cents, status, created_at from users where id = $1', [payload.sub]);
    const user = rows[0];
    if (!user || user.status !== 'active') return reply.code(401).send({ error: 'invalid_user' });
    req.user = user;
  } catch {
    return reply.code(401).send({ error: 'invalid_token' });
  }
}

async function authApiKey(req, reply) {
  const raw = parseBearer(req.headers.authorization);
  if (!raw) return reply.code(401).send({ error: 'missing_api_key' });
  const keyHash = sha256(raw);
  const { rows } = await pool.query(
    `select ak.id as api_key_id, u.id as user_id, u.email, u.role, u.balance_cents, u.status
       from api_keys ak
       join users u on u.id = ak.user_id
      where ak.key_hash = $1 and ak.status = 'active'`,
    [keyHash]
  );
  const row = rows[0];
  if (!row || row.status !== 'active') return reply.code(401).send({ error: 'invalid_api_key' });
  req.apiUser = row;
}

function requireAdmin(req, reply) {
  if (req.user?.role !== 'super_admin') {
    return reply.code(403).send({ error: 'admin_only' });
  }
}

async function addLedger(client, { userId, type, amountCents, referenceType = null, referenceId = null, note = null }) {
  const { rows } = await client.query('select balance_cents from users where id = $1', [userId]);
  const balanceAfter = Number(rows[0].balance_cents);
  await client.query(
    `insert into balance_ledger (id, user_id, type, amount_cents, balance_after_cents, reference_type, reference_id, note)
     values ($1,$2,$3,$4,$5,$6,$7,$8)`,
    [genId(), userId, type, amountCents, balanceAfter, referenceType, referenceId, note]
  );
}

app.get('/health', async () => ({ ok: true }));
app.get('/', async () => ({ name: 'CLIAPI', version: '0.1.0', docs: '/health' }));

app.post('/auth/register', async (req, reply) => {
  if (!ALLOW_REGISTRATION) return reply.code(403).send({ error: 'registration_disabled' });
  const { email, password } = req.body || {};
  if (!email || !password) return reply.code(400).send({ error: 'email_and_password_required' });
  const passwordHash = await bcrypt.hash(password, 10);
  try {
    const { rows } = await pool.query(
      `insert into users (id, email, password_hash, role, status, balance_cents)
       values ($1,$2,$3,'user','active',0)
       returning id, email, role, balance_cents, created_at`,
      [genId(), String(email).toLowerCase(), passwordHash]
    );
    const user = rows[0];
    return { token: signUser(user), user };
  } catch (error) {
    if (String(error.message).includes('users_email_key')) return reply.code(409).send({ error: 'email_exists' });
    throw error;
  }
});

app.post('/auth/login', async (req, reply) => {
  const { email, password } = req.body || {};
  if (!email || !password) return reply.code(400).send({ error: 'email_and_password_required' });
  const { rows } = await pool.query('select id, email, password_hash, role, status, balance_cents, created_at from users where email = $1', [String(email).toLowerCase()]);
  const user = rows[0];
  if (!user || user.status !== 'active') return reply.code(401).send({ error: 'invalid_credentials' });
  const ok = await bcrypt.compare(password, user.password_hash);
  if (!ok) return reply.code(401).send({ error: 'invalid_credentials' });
  return { token: signUser(user), user: { id: user.id, email: user.email, role: user.role, balance_cents: user.balance_cents, created_at: user.created_at } };
});

app.get('/me', { preHandler: authUser }, async (req) => ({ user: req.user }));
app.get('/me/balance', { preHandler: authUser }, async (req) => ({ balance_cents: Number(req.user.balance_cents) }));
app.get('/me/ledger', { preHandler: authUser }, async (req) => {
  const { rows } = await pool.query('select * from balance_ledger where user_id = $1 order by created_at desc limit 100', [req.user.id]);
  return { items: rows };
});
app.get('/me/usage', { preHandler: authUser }, async (req) => {
  const { rows } = await pool.query('select * from token_usage_logs where user_id = $1 order by created_at desc limit 100', [req.user.id]);
  return { items: rows };
});

app.post('/me/api-keys', { preHandler: authUser }, async (req) => {
  const { name } = req.body || {};
  const secret = `${API_KEY_PREFIX}_${crypto.randomBytes(24).toString('hex')}`;
  const prefix = secret.slice(0, 12);
  await pool.query(
    `insert into api_keys (id, user_id, name, key_prefix, key_hash, status)
     values ($1,$2,$3,$4,$5,'active')`,
    [genId(), req.user.id, name || 'default', prefix, sha256(secret)]
  );
  return { api_key: secret, key_prefix: prefix };
});

app.get('/me/api-keys', { preHandler: authUser }, async (req) => {
  const { rows } = await pool.query('select id, name, key_prefix, status, last_used_at, created_at from api_keys where user_id = $1 order by created_at desc', [req.user.id]);
  return { items: rows };
});

app.delete('/me/api-keys/:id', { preHandler: authUser }, async (req) => {
  await pool.query('update api_keys set status = $1 where id = $2 and user_id = $3', ['revoked', req.params.id, req.user.id]);
  return { ok: true };
});

app.post('/me/redeem', { preHandler: authUser }, async (req, reply) => {
  const { code } = req.body || {};
  if (!code) return reply.code(400).send({ error: 'code_required' });
  const client = await pool.connect();
  try {
    await client.query('begin');
    const { rows } = await client.query('select * from redemption_codes where code = $1 for update', [String(code).trim()]);
    const item = rows[0];
    if (!item) throw new Error('invalid_code');
    if (item.status !== 'unused') throw new Error('code_not_available');
    if (item.expires_at && new Date(item.expires_at).getTime() < Date.now()) throw new Error('code_expired');

    await client.query(
      `update redemption_codes
          set status = 'used', used_by = $1, used_at = now(), used_count = used_count + 1
        where id = $2`,
      [req.user.id, item.id]
    );
    await client.query('update users set balance_cents = balance_cents + $1, updated_at = now() where id = $2', [item.amount_cents, req.user.id]);
    await addLedger(client, {
      userId: req.user.id,
      type: 'recharge_code',
      amountCents: Number(item.amount_cents),
      referenceType: 'redemption_code',
      referenceId: item.id,
      note: `兑换码充值 ${code}`
    });
    await client.query('commit');
    const { rows: userRows } = await pool.query('select balance_cents from users where id = $1', [req.user.id]);
    return { ok: true, credited_cents: Number(item.amount_cents), balance_cents: Number(userRows[0].balance_cents) };
  } catch (error) {
    await client.query('rollback');
    const map = new Set(['invalid_code', 'code_not_available', 'code_expired']);
    if (map.has(error.message)) return reply.code(400).send({ error: error.message });
    throw error;
  } finally {
    client.release();
  }
});

app.post('/admin/redemption-codes', { preHandler: [authUser, requireAdmin] }, async (req, reply) => {
  const { amount_cents, count = 1, expires_at = null } = req.body || {};
  if (!amount_cents || Number(amount_cents) <= 0) return reply.code(400).send({ error: 'amount_cents_required' });
  const codes = [];
  for (let i = 0; i < Number(count); i += 1) {
    const code = `TOPUP-${crypto.randomBytes(3).toString('hex').toUpperCase()}-${crypto.randomBytes(3).toString('hex').toUpperCase()}`;
    await pool.query(
      `insert into redemption_codes (id, code, amount_cents, status, created_by, expires_at, max_uses, used_count)
       values ($1,$2,$3,'unused',$4,$5,1,0)`,
      [genId(), code, Number(amount_cents), req.user.id, expires_at]
    );
    codes.push(code);
  }
  return { codes };
});

app.get('/admin/users', { preHandler: [authUser, requireAdmin] }, async () => {
  const { rows } = await pool.query('select id, email, role, status, balance_cents, created_at from users order by created_at desc limit 200');
  return { items: rows };
});

app.post('/admin/users/:id/adjust-balance', { preHandler: [authUser, requireAdmin] }, async (req, reply) => {
  const { amount_cents, note } = req.body || {};
  if (!Number.isInteger(amount_cents)) return reply.code(400).send({ error: 'amount_cents_must_be_integer' });
  const client = await pool.connect();
  try {
    await client.query('begin');
    await client.query('update users set balance_cents = balance_cents + $1, updated_at = now() where id = $2', [amount_cents, req.params.id]);
    await addLedger(client, { userId: req.params.id, type: 'admin_adjust', amountCents: amount_cents, referenceType: 'admin', referenceId: req.user.id, note: note || 'admin adjust' });
    await client.query('commit');
    return { ok: true };
  } catch (error) {
    await client.query('rollback');
    throw error;
  } finally {
    client.release();
  }
});

app.post('/v1/chat/completions', { preHandler: authApiKey }, async (req, reply) => {
  if (!OPENAI_API_KEY) return reply.code(500).send({ error: 'upstream_not_configured' });
  if (Number(req.apiUser.balance_cents) <= 0) return reply.code(402).send({ error: 'insufficient_balance' });

  const upstreamRes = await fetch(`${OPENAI_BASE_URL}/chat/completions`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      authorization: `Bearer ${OPENAI_API_KEY}`
    },
    body: JSON.stringify(req.body)
  });

  const data = await upstreamRes.json();
  if (!upstreamRes.ok) return reply.code(upstreamRes.status).send(data);

  const usage = data.usage || {};
  const inputTokens = Number(usage.prompt_tokens || usage.input_tokens || 0);
  const outputTokens = Number(usage.completion_tokens || usage.output_tokens || 0);
  const totalTokens = Number(usage.total_tokens || (inputTokens + outputTokens));
  const cents = costCents(totalTokens);

  const client = await pool.connect();
  try {
    await client.query('begin');
    const userRes = await client.query('select balance_cents from users where id = $1 for update', [req.apiUser.user_id]);
    const currentBalance = Number(userRes.rows[0].balance_cents);
    if (currentBalance < cents) {
      await client.query('rollback');
      return reply.code(402).send({ error: 'insufficient_balance_after_usage', required_cents: cents, balance_cents: currentBalance });
    }
    const usageId = genId();
    await client.query(
      `insert into token_usage_logs
      (id, user_id, api_key_id, model, input_tokens, output_tokens, total_tokens, cost_cents, request_id)
      values ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
      [usageId, req.apiUser.user_id, req.apiUser.api_key_id, data.model || req.body.model || 'unknown', inputTokens, outputTokens, totalTokens, cents, data.id || null]
    );
    await client.query('update users set balance_cents = balance_cents - $1, updated_at = now() where id = $2', [cents, req.apiUser.user_id]);
    await addLedger(client, {
      userId: req.apiUser.user_id,
      type: 'usage_deduct',
      amountCents: -cents,
      referenceType: 'usage_log',
      referenceId: usageId,
      note: `模型调用 ${data.model || req.body.model || 'unknown'} / ${totalTokens} tokens`
    });
    await client.query('update api_keys set last_used_at = now() where id = $1', [req.apiUser.api_key_id]);
    await client.query('commit');
  } catch (error) {
    await client.query('rollback');
    throw error;
  } finally {
    client.release();
  }

  reply.header('x-proxy-billed-cents', String(cents));
  return data;
});

app.listen({ host: '0.0.0.0', port: PORT })
  .then(() => app.log.info(`CLIAPI listening on ${PORT}`))
  .catch((err) => {
    app.log.error(err);
    process.exit(1);
  });
