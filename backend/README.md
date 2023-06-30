`api` in the root is special, vercel will detect the runtime and generate a standalone go module called `api`.

Every file not begins with `_` will be treated as a standalone serverless function, compiled into a standalone go module.

So, in `api` package, we can not refer `internal`, since it is another module.
