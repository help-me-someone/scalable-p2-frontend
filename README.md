# Toktik frontend
Help.

# Build
```sh
make build
```
# Build and Run
```sh
make run
```

# Developer Notes
Hot reloading for templates is somewhat supported though very limited.

Requirements:
- Tailwind must be running on watch mode.
- The `debug` value in `main.go` must be set to true.

Ypu can run tailwind on watch mode using the following command.
```sh
npm run watch
```

Now in another terminal, you can run `make run` and hot reloading will be on. (Be mindful of browser caching)
