# Proto buf stub generator
    - https://buf.build/docs/cli/

## Stub generator
    - docker compose up buf_build_gateway_proto

## Breaking changes checker
    - docker run -v $(pwd):/statistic:rw -w /statistic bufbuild/buf:latest breaking --against 'https://github.com/foo/bar.git#main'

## Linters
    - docker compose up buf_lint_gateway_proto

## Formatter
    - docker compose up buf_format_gateway_proto

## Dependencies update
    - docker compose up buf_dep_update_gateway_proto

## GO module init
    - docker compose up module_init_gateway_proto
