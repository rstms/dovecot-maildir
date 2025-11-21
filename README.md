# dovecot-maildir

Utility for viewing and manipulating Maildir files maintained by a dovecot
IMAP server.  Can uncompress messages compressed with zstandard.
IMPORTANT: Designed to be run with the dovecot daemon stopped, as it modifies
maildir files in place without use of locking or indexing mechanisms.
