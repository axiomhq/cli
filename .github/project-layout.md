# Project Layout

This document describes the project layout of this repository. It is derived
from the [Standard Go Project Layout][1].

Additional background information is available at [Go Project Layout][2].

  [1]: https://github.com/golang-standards/project-layout
  [2]: https://medium.com/golang-learn/go-project-layout-e5213cdcfaa2

## Go Directories

### `/cmd`

Main applications for this project.

The directory name for each application should match the name of the executable
(e.g., `/cmd/myapp`).

If code can be imported and used in other projects, then it should live in the
`/pkg` directory. If the code is not or should not be reusable, that code should
live in the `/internal` directory.

### `/internal`

Private application and library code.

### `/pkg`

Library code that's safe to use by external applications (e.g.,
`/pkg/mypubliclib`). Other projects will import these libraries expecting them
to work!

## Other Directories

### `/tools`

Supporting tools for this project. Note that these tools can import code from
the `/pkg` and `/internal` directories.
