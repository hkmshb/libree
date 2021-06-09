# libree
A collection of scripts for managing my digital library assets across various cloud storage
providers.

## Database Setup
To setup the `libree` CouchDB database run:

```sh
$ grunt
```

This uses `Grunt` with the `grunt-couchdb` plugin to create and setup the database if it doesn't
already exist using contents from `couchdb/bootstrap`.

The under listed environment variables are required for the database creation. Create a `.env`
file with these variables.

```sh
// file: .env
export LIBREE_COUCHDB_URL=http://localhost:5984
export LIBREE_COUCHDB_USER=
export LIBREE_COUCHDB_PASS=
```

## Libree Data Management
Use the `libree/main.go` script to manage records within the `libree` database. Here is the
output for `--help` flag for the script:

```sh
$ run go libree/main.go --help

Usage: libree [OPTIONS] ...
  -d string
        Path to directory to process
  -u string
        Target API service url (default "http://localhost:5984/libree")
```
