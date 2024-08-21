# gogetgreen

## About

> gogetgreen is an awkward pun. It's not about saving the environment, sorry. I nearly called it gogreen but that seemed too much.

gogetgreen is a test runner for Go (aka golang), the programming language. It's a desktop app which can watch your project for file changes.

It's early days. It's not very good yet.

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on <http://localhost:34115>. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.
