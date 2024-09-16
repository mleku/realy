relayer basic
=============

  - a basic relay implementation based on relayer.
  - uses postgres, which I think must be over version 12 since it uses generated columns.
  - it has some antispam limits, tries to delete old stuff so things don't get out of control, and some other small optimizations.

running
-------

grab a binary from the releases page and run it with the environment variable POSTGRESQL_DATABASE set to some postgres url:

    POSTGRESQL_DATABASE=postgres://name:pass@localhost:5432/dbname ./realy-basic

it also accepts a HOST and a PORT environment variables.

compiling
---------

if you know Go you already know this:

    go install mleku.dev/realy@latest

or something like that.
