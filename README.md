# go ionic server

Translate [this](https://github.com/gsrai/Ionic) from `typescript` & `node` to `go`.

The node version has the following dependencies:

```json
"devDependencies": {
  "@types/express": "^4.17.13",
  "@types/json2csv": "^5.0.3",
  "@types/node": "^16.11.13",
  "@typescript-eslint/eslint-plugin": "^5.8.0",
  "@typescript-eslint/parser": "^5.8.0",
  "eslint": "^8.5.0",
  "eslint-config-prettier": "^8.3.0",
  "eslint-plugin-prettier": "^4.0.0",
  "nodemon": "^2.0.14",
  "pino-pretty": "^7.3.0",
  "prettier": "2.5.1",
  "ts-node": "^10.4.0",
  "typescript": "^4.5.4"
},
"dependencies": {
  "async-sema": "^3.1.1",
  "axios": "^0.24.0",
  "csv-parse": "^5.0.3",
  "dotenv": "^10.0.0",
  "express": "^4.17.1",
  "json2csv": "^5.0.6",
  "pino": "^7.5.1",
  "pino-http": "^6.4.0"
}
```

Go comes with a logger, http client and server, formatter, types

```sh
# Test Endpoints
# Read input CSV
curl http://localhost:8080/input/load
# Fetch block heights
curl http://localhost:8080/block/heights
```

javascript version took: 27.69 seconds
go version took: 19.23 seconds
