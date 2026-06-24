# Proto buf stub generator
    - https://buf.build/docs/cli/

## Stub generator
    - docker compose up buf_build_report_proto

## Breaking changes checker
    - docker run -v $(pwd):/statistic:rw -w /statistic bufbuild/buf:latest breaking --against 'https://github.com/foo/bar.git#main'

## Linters
    - docker compose up buf_lint_report_proto

## Formatter
    - docker compose up buf_format_report_proto

## Dependencies update
    - docker compose up buf_dep_update_report_proto

## GO module init
    - docker compose up module_init_report_proto
