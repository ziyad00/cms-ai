export function goApiBaseUrl() {
  return process.env.GO_API_BASE_URL || 'http://127.0.0.1:8081'
}

export function joinUrl(base, path) {
  if (!base) throw new Error('base is required')
  if (!path) throw new Error('path is required')
  const b = base.endsWith('/') ? base.slice(0, -1) : base
  const p = path.startsWith('/') ? path : `/${path}`
  return `${b}${p}`
}

export async function postJSON(path, body, { baseUrl = goApiBaseUrl(), headers = {} } = {}) {
  const url = joinUrl(baseUrl, path)

  // Force external HTTP request for server-side calls using node:http
  if (typeof window === 'undefined') {
    const http = await import('node:http')
    const { URL } = await import('node:url')

    return new Promise((resolve, reject) => {
      const parsedUrl = new URL(url)
      const postData = JSON.stringify(body)

      const options = {
        hostname: parsedUrl.hostname,
        port: parsedUrl.port || (parsedUrl.protocol === 'https:' ? 443 : 80),
        path: parsedUrl.pathname + parsedUrl.search,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
          'Host': parsedUrl.hostname + (parsedUrl.port ? ':' + parsedUrl.port : ''),
          ...headers
        }
      }

      const req = http.request(options, (res) => {
        let data = ''
        res.on('data', (chunk) => { data += chunk })
        res.on('end', () => {
          let parsed
          try {
            parsed = data ? JSON.parse(data) : null
          } catch {
            parsed = { raw: data }
          }
          resolve({ status: res.statusCode, body: parsed })
        })
      })

      req.on('error', (error) => {
        reject(error)
      })

      req.setTimeout(10000, () => {
        req.destroy()
        reject(new Error('Request timeout'))
      })

      req.write(postData)
      req.end()
    })
  }

  // Client-side fallback to fetch
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...headers },
    body: JSON.stringify(body),
  })

  const text = await res.text()
  let parsed
  try {
    parsed = text ? JSON.parse(text) : null
  } catch {
    parsed = { raw: text }
  }

  return { status: res.status, body: parsed }
}

export async function getJSON(path, { baseUrl = goApiBaseUrl(), headers = {} } = {}) {
  const url = joinUrl(baseUrl, path)

  // Force external HTTP request for server-side calls using node:http
  if (typeof window === 'undefined') {
    const http = await import('node:http')
    const { URL } = await import('node:url')

    return new Promise((resolve, reject) => {
      const parsedUrl = new URL(url)

      const options = {
        hostname: parsedUrl.hostname,
        port: parsedUrl.port || (parsedUrl.protocol === 'https:' ? 443 : 80),
        path: parsedUrl.pathname + parsedUrl.search,
        method: 'GET',
        headers: {
          'Accept': 'application/json',
          'Host': parsedUrl.hostname + (parsedUrl.port ? ':' + parsedUrl.port : ''),
          ...headers
        }
      }

      const req = http.request(options, (res) => {
        let data = ''
        res.on('data', (chunk) => { data += chunk })
        res.on('end', () => {
          let parsed
          try {
            parsed = data ? JSON.parse(data) : null
          } catch {
            parsed = { raw: data }
          }
          resolve({ status: res.statusCode, body: parsed })
        })
      })

      req.on('error', (error) => {
        reject(error)
      })

      req.setTimeout(10000, () => {
        req.destroy()
        reject(new Error('Request timeout'))
      })

      req.end()
    })
  }

  // Client-side fallback to fetch
  const res = await fetch(url, {
    method: 'GET',
    headers: { 'Accept': 'application/json', ...headers },
  })

  const text = await res.text()
  let parsed
  try {
    parsed = text ? JSON.parse(text) : null
  } catch {
    parsed = { raw: text }
  }

  return { status: res.status, body: parsed }
}
