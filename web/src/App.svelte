<script lang="ts">
  import {
    startRegistration,
    startAuthentication,
  } from '@simplewebauthn/browser'

  type Mode = 'register' | 'login'

  type Status =
    | { kind: 'idle' }
    | { kind: 'working' }
    | { kind: 'ok'; username: string; mode: Mode }
    | { kind: 'error'; message: string }

  let mode = $state<Mode>('register')
  let username = $state('')
  let email = $state('')
  let status = $state<Status>({ kind: 'idle' })

  function switchMode(next: Mode) {
    mode = next
    status = { kind: 'idle' }
  }

  async function register() {
    status = { kind: 'working' }

    try {
      const begin = await fetch('/api/auth/register/begin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email }),
      })

      if (!begin.ok) {
        throw new Error(await begin.text() || `begin: ${begin.status}`)
      }

      const { sessionId, options } = await begin.json()

      // browser prompts for face/finger/pin here. options.publicKey
      // is the WebAuthn options the Go server built; the
      // simplewebauthn helper handles the base64/ArrayBuffer dance.
      const credential = await startRegistration(options.publicKey)

      const finish = await fetch('/api/auth/register/finish', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Pixeltown-Session': sessionId,
        },
        body: JSON.stringify(credential),
      })

      if (!finish.ok) {
        throw new Error(await finish.text() || `finish: ${finish.status}`)
      }

      const { user } = await finish.json()
      status = { kind: 'ok', username: user.username, mode: 'register' }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : String(e)
      status = { kind: 'error', message }
    }
  }

  async function login() {
    status = { kind: 'working' }

    try {
      const begin = await fetch('/api/auth/login/begin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      })

      if (!begin.ok) {
        throw new Error(await begin.text() || `begin: ${begin.status}`)
      }

      const { sessionId, options } = await begin.json()

      const credential = await startAuthentication(options.publicKey)

      const finish = await fetch('/api/auth/login/finish', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Pixeltown-Session': sessionId,
        },
        body: JSON.stringify(credential),
      })

      if (!finish.ok) {
        throw new Error(await finish.text() || `finish: ${finish.status}`)
      }

      const { user } = await finish.json()
      status = { kind: 'ok', username: user.username, mode: 'login' }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : String(e)
      status = { kind: 'error', message }
    }
  }
</script>

<main>
  <h1>pixeltown</h1>

  {#if status.kind === 'ok'}
    <p>
      {status.mode === 'register' ? 'signed up' : 'signed in'} as
      <strong>{status.username}</strong>.
    </p>
  {:else}
    <form
      onsubmit={(e) => {
        e.preventDefault()
        if (mode === 'register') register()
        else login()
      }}
    >
      {#if mode === 'register'}
        <label>
          username
          <input bind:value={username} required maxlength="32" />
        </label>
      {/if}

      <label>
        email
        <input type="email" bind:value={email} required />
      </label>

      <button type="submit" disabled={status.kind === 'working'}>
        {#if status.kind === 'working'}
          {mode === 'register' ? 'creating passkey...' : 'signing in...'}
        {:else}
          {mode === 'register' ? 'sign up' : 'sign in'}
        {/if}
      </button>

      {#if status.kind === 'error'}
        <p class="error">{status.message}</p>
      {/if}
    </form>

    <p class="switch">
      {#if mode === 'register'}
        already have an account?
        <button type="button" onclick={() => switchMode('login')}>
          log in
        </button>
      {:else}
        need an account?
        <button type="button" onclick={() => switchMode('register')}>
          sign up
        </button>
      {/if}
    </p>
  {/if}
</main>

<style>
  main {
    max-width: 24rem;
    margin: 4rem auto;
    padding: 0 1rem;
    font-family: system-ui, sans-serif;
  }
  h1 {
    font-size: 1.5rem;
    margin-bottom: 1.25rem;
  }
  form {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  label {
    display: flex;
    flex-direction: column;
    font-size: 0.85rem;
    gap: 0.25rem;
  }
  input {
    padding: 0.4rem 0.5rem;
    font-size: 1rem;
  }
  button {
    padding: 0.5rem;
    font-size: 1rem;
    cursor: pointer;
  }
  button:disabled {
    opacity: 0.6;
    cursor: default;
  }
  .error {
    color: #b00020;
    font-size: 0.85rem;
  }
  .switch {
    margin-top: 1.5rem;
    font-size: 0.85rem;
  }
  .switch button {
    padding: 0;
    font-size: 0.85rem;
    background: none;
    border: none;
    color: #aa3bff;
    text-decoration: underline;
    cursor: pointer;
  }
</style>
