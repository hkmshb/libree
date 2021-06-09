require('dotenv').config();

const couchUrl = process.env.LIBREE_COUCHDB_URL;
const username = process.env.LIBREE_COUCHDB_USER;
const password = process.env.LIBREE_COUCHDB_PASS;

module.exports = function (grunt) {
  grunt.initConfig({
    pkg: grunt.file.readJSON('package.json'),
    couchdb: {
      url: couchUrl,
      user: username,
      password: password,
      bootstrap: {dir: 'couchdb/bootstrap'}
    }
  });

  grunt.loadNpmTasks('grunt-couchdb');
  grunt.registerTask('default', ['couchdb:bootstrap']);
}
