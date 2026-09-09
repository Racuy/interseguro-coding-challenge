// config.js sets window.APP_CONFIG.apiBase at container start, falls back to same-host:8080 for local docker compose
const API_BASE = (window.APP_CONFIG && window.APP_CONFIG.apiBase) || `http://${window.location.hostname}:8080`

let token = null
let rows = 3
let cols = 3
let matrixValues = defaultMatrix(rows, cols)

const loginView = document.getElementById('login-view')
const appView = document.getElementById('app-view')
const loginForm = document.getElementById('login-form')
const loginError = document.getElementById('login-error')
const loginBtn = document.getElementById('login-btn')
const logoutBtn = document.getElementById('logout-btn')

const layout = document.getElementById('layout')
const rowsInput = document.getElementById('rows-input')
const colsInput = document.getElementById('cols-input')
const matrixGrid = document.getElementById('matrix-grid')
const randomBtn = document.getElementById('random-btn')
const clearBtn = document.getElementById('clear-btn')
const submitBtn = document.getElementById('submit-btn')
const submitError = document.getElementById('submit-error')

const rotatedGrid = document.getElementById('rotated-grid')
const qGrid = document.getElementById('q-grid')
const rGrid = document.getElementById('r-grid')

// starts empty, the user types their own values, nothing pre-filled
function defaultMatrix(r, c) {
  const data = []
  for (let i = 0; i < r; i++) data.push(new Array(c).fill(''))
  return data
}

// blanks become 0 only at submit time, this is the only rounding anywhere since neither backend rounds
function toNumericMatrix(values) {
  return values.map((row) => row.map((v) => (v === '' ? 0 : round3(Number(v)))))
}

function round3(value) {
  return Math.round(value * 1000) / 1000
}

// trims decimals past 3 digits only, leaves a mid-edit string like a lone minus sign or trailing dot alone
function limitToThreeDecimals(rawValue) {
  const parts = rawValue.split('.')
  if (parts.length < 2 || parts[1].length <= 3) return rawValue
  return `${parts[0]}.${parts[1].slice(0, 3)}`
}

function setLoading(button, loading) {
  button.disabled = loading
  button.querySelector('.spinner').hidden = !loading
}

function showError(el, message) {
  el.textContent = message
  el.hidden = false
}

function hideError(el) {
  el.hidden = true
}

async function login(username, password) {
  const res = await fetch(`${API_BASE}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'no se pudo iniciar sesión')
  return data.token
}

async function processMatrix(matrix) {
  const res = await fetch(`${API_BASE}/api/v1/matrix/process`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ matrix }),
  })
  const data = await res.json()
  if (res.status === 401) {
    logout()
    throw new Error('tu sesión expiró, iniciá sesión de nuevo')
  }
  if (!res.ok) throw new Error(data.error || 'no se pudo procesar la matriz')
  return data
}

function renderMatrixGrid() {
  matrixGrid.style.gridTemplateColumns = `repeat(${cols}, 1fr)`
  matrixGrid.innerHTML = ''

  for (let i = 0; i < rows; i++) {
    for (let j = 0; j < cols; j++) {
      const input = document.createElement('input')
      input.type = 'number'
      input.step = '0.001'
      input.placeholder = '0'
      input.value = matrixValues[i]?.[j] ?? ''
      input.dataset.row = String(i)
      input.dataset.col = String(j)
      input.addEventListener('input', () => {
        const limited = limitToThreeDecimals(input.value)
        if (limited !== input.value) input.value = limited
        matrixValues[i][j] = limited
      })
      matrixGrid.appendChild(input)
    }
  }
}

function resizeMatrix(newRows, newCols) {
  const resized = []
  for (let i = 0; i < newRows; i++) {
    const row = []
    for (let j = 0; j < newCols; j++) row.push(matrixValues[i]?.[j] ?? '')
    resized.push(row)
  }
  matrixValues = resized
  rows = newRows
  cols = newCols
  renderMatrixGrid()
}

function renderReadonlyGrid(container, data) {
  container.style.gridTemplateColumns = `repeat(${data[0].length}, 1fr)`
  container.innerHTML = ''
  for (const row of data) {
    for (const value of row) {
      const cell = document.createElement('div')
      cell.className = 'cell'
      cell.textContent = formatNumber(value)
      cell.title = String(value)
      container.appendChild(cell)
    }
  }
}

// display only: rounds to 3 decimals then lets String() drop trailing zeros (-1.000 becomes "-1"), -0 always shows as plain 0
function formatNumber(value) {
  const rounded = Math.round(value * 1000) / 1000
  return String(rounded === 0 ? 0 : rounded)
}

function renderResults(result) {
  renderReadonlyGrid(rotatedGrid, result.rotated)
  renderReadonlyGrid(qGrid, result.q)
  renderReadonlyGrid(rGrid, result.r)

  document.getElementById('stat-max').textContent = formatNumber(result.stats.max)
  document.getElementById('stat-min').textContent = formatNumber(result.stats.min)
  document.getElementById('stat-avg').textContent = formatNumber(result.stats.average)
  document.getElementById('stat-sum').textContent = formatNumber(result.stats.sum)

  const badge = document.getElementById('stat-diagonal')
  const diagonalNames = (result.stats.diagonalMatrices || []).map((m) => m.toUpperCase()).join(', ')
  badge.textContent = result.stats.isDiagonal ? `Sí (${diagonalNames})` : 'No'
  badge.className = `badge ${result.stats.isDiagonal ? 'badge-true' : 'badge-false'}`

  layout.classList.add('has-results')
}

function showAppView() {
  loginView.hidden = true
  appView.hidden = false
  renderMatrixGrid()
}

function logout() {
  token = null
  rows = 3
  cols = 3
  matrixValues = defaultMatrix(rows, cols)
  rowsInput.value = rows
  colsInput.value = cols
  layout.classList.remove('has-results')
  appView.hidden = true
  loginView.hidden = false
  loginForm.reset()
}

loginForm.addEventListener('submit', async (event) => {
  event.preventDefault()
  hideError(loginError)
  setLoading(loginBtn, true)

  const username = document.getElementById('username').value
  const password = document.getElementById('password').value

  try {
    token = await login(username, password)
    showAppView()
  } catch (err) {
    showError(loginError, err.message)
  } finally {
    setLoading(loginBtn, false)
  }
})

logoutBtn.addEventListener('click', logout)

rowsInput.addEventListener('change', () => {
  const value = Math.min(12, Math.max(1, Number(rowsInput.value) || 1))
  rowsInput.value = value
  resizeMatrix(value, cols)
})

colsInput.addEventListener('change', () => {
  const value = Math.min(12, Math.max(1, Number(colsInput.value) || 1))
  colsInput.value = value
  resizeMatrix(rows, value)
})

randomBtn.addEventListener('click', () => {
  matrixValues = matrixValues.map((row) => row.map(() => round3(Math.random() * 20 - 10)))
  renderMatrixGrid()
})

clearBtn.addEventListener('click', () => {
  matrixValues = matrixValues.map((row) => row.map(() => ''))
  renderMatrixGrid()
})

submitBtn.addEventListener('click', async () => {
  hideError(submitError)
  setLoading(submitBtn, true)

  try {
    const result = await processMatrix(toNumericMatrix(matrixValues))
    renderResults(result)
  } catch (err) {
    showError(submitError, err.message)
  } finally {
    setLoading(submitBtn, false)
  }
})
