#!/bin/bash

sshpass -p ensure17 ssh ensure@134.130.169.26 "cqlsh -u cassandra -p cassandra -e 'TRUNCATE TABLE clonemap.logging_app'"
sshpass -p ensure17 ssh ensure@134.130.169.26 "cqlsh -u cassandra -p cassandra -e 'TRUNCATE TABLE clonemap.logging_msg'"
sshpass -p ensure17 ssh ensure@134.130.169.26 "cqlsh -u cassandra -p cassandra -e 'TRUNCATE TABLE clonemap.logging_error'"
sshpass -p ensure17 ssh ensure@134.130.169.26 "cqlsh -u cassandra -p cassandra -e 'TRUNCATE TABLE clonemap.logging_status'"
sshpass -p ensure17 ssh ensure@134.130.169.26 "cqlsh -u cassandra -p cassandra -e 'TRUNCATE TABLE clonemap.logging_debug'"
sshpass -p ensure17 ssh ensure@134.130.169.26 "cqlsh -u cassandra -p cassandra -e 'TRUNCATE TABLE clonemap.state'"