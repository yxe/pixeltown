<script lang="ts">
  import { startRegistration } from '@simplewebauthn/browser'

  type Status =
    | { kind: 'idle' }
    | { kind: 'working' }
    | { kind: 'ok'; username: string }
    | { kind: 'error'; message: string }

  let username = $state('')
  let email = $state('')
  let status = $state<Status>({ kind: 'idle' })

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
      status = { kind: 'ok', username: user.username }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : String(e)
      status = { kind: 'error', message }
    }
  }
</script>

<main>
  <h1>pixeltown</h1>

  {#if status.kind === 'ok'}
    <p>signed up as <strong>{status.username}</strong>.</p>
  {:else}
    <form
      onsubmit={(e) => {
        e.preventDefault()
        register()
      }}
    >
      <label>
        username
        <input bind:value={username} required maxlength="32" />
      </label>

      <label>
        email
        <input type="email" bind:value={email} required />
      </label>

      <button type="submit" disabled={status.kind === 'working'}>
        {status.kind === 'working' ? 'creating passkey...' : 'sign up'}
      </button>

      {#if status.kind === 'error'}
        <p class="error">{status.message}</p>
      {/if}
    </form>
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
</style>
