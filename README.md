# Reto Técnico Interseguro

Dos APIs que se comunican por HTTP:

- **go-api** (Go + Fiber): recibe una matriz, la rota 90° y calcula su factorización QR sobre la matriz original.
- **node-api** (Node + Express): recibe Q y R y calcula estadísticas sobre sus valores.

Hay también un frontend simple para probar todo desde el navegador.

## Cómo levantarlo

```bash
docker compose up --build
```

- Frontend: http://localhost:3001
- go-api: http://localhost:8080
- node-api: sin puerto expuesto al host, solo accesible dentro de la red de Docker

## Cómo funciona

1. El frontend envía la matriz a `POST /api/v1/matrix/process` de go-api.
2. go-api rota la matriz 90° y calcula la QR de la matriz original.
3. go-api llama a `POST /api/v1/stats` de node-api enviando solo `q` y `r`.
4. node-api devuelve máximo, mínimo, promedio, suma y si alguna matriz es diagonal.
5. go-api responde al frontend con `original`, `rotated`, `q`, `r` y `stats`.

Ejemplo:

```bash
curl -X POST http://localhost:8080/api/v1/matrix/process \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[1,2,3],[4,5,6]]}'
```

## Decisiones que tomé

El enunciado deja algunas cosas abiertas, así que las resolví así:

- **La QR se calcula sobre la matriz original, no sobre la rotada.** El PDF pide "la factorización QR de dicha matriz" refiriéndose a la de entrada. La rotación se devuelve como dato adicional.
- **Las estadísticas se calculan sobre Q y R juntas**, como un solo conjunto de valores, porque el PDF habla de "todos los valores de las matrices".
- **La verificación de diagonal es por matriz**: da `true` si Q o R es diagonal, y la respuesta indica cuál.
- **Los backends no redondean.** Devuelven los números completos y el frontend los muestra con 3 decimales.
- **La diagonal de R siempre es positiva.** La QR no es única (se pueden invertir signos de una columna de Q y su fila de R), así que aplico esa convención para que el resultado sea siempre el mismo.

## Tests

```bash
(cd go-api && go test ./...)
(cd node-api && npm test)
```

Los tests de go-api verifican que Q·R reconstruye la matriz y que Q es ortogonal, incluyendo matrices rectangulares y con columnas en cero. Los de node-api cubren las estadísticas y la detección de diagonal. Hay además un test de integración entre las dos APIs.