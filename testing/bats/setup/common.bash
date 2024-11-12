load setup/windows-compat
load setup/query-server-common

if [ -z "$BATS_TMPDIR" ]; then
    export BATS_TMPDIR=$HOME/batstmp/
    mkdir $BATS_TMPDIR
fi

nativebatsdir() { echo `nativepath $BATS_TEST_DIRNAME/$1`; }
batshelper() { echo `nativebatsdir helper/$1`; }

setup_common() {
    run psql --version
    if [[ ! "$output" =~ "(PostgreSQL) 15" ]] && [[ ! "$output" =~ "(PostgreSQL) 16" ]]; then
        echo "PSQL must be version 15, got $output"
        return 1
    fi

    export PATH=$PATH:~/go/bin
    cd $BATS_TMPDIR

    # remove directory if exists
    # reruns recycle pids
    rm -rf "dolt-repo-$$"

    # Append the directory name with the pid of the calling process so
    # multiple tests can be run in parallel on the same machine
    mkdir "dolt-repo-$$"
    cd "dolt-repo-$$"
    nativevar DOLTGRES_DATA_DIR "$(pwd)" /p

    if [ -z "$DOLT_TEST_RETRIES" ]; then
        export BATS_TEST_RETRIES="$DOLT_TEST_RETRIES"
    fi
}

teardown_common() {
    # rm -rf can fail with a "directory not empty" error in some cases. This seems to be a misleading
    # error message; the real error is that a file is still in use. Instead of waiting longer for
    # any processes to finish, we just ignore any error removing temp files and use 'true' as the last
    # command in this function to ensure that teardown_common doesn't fail a test just because we
    # couldn't delete any temporary test files.
    stop_sql_server
    rm -rf "$BATS_TMPDIR/dolt-repo-$$"
    true
}

query_server() {
    nativevar PGPASSWORD "password" /w
    psql -U "${SQL_USER:-postgres}" -h localhost -p $PORT "$@" postgres
}

query_server_for_db() {
    nativevar PGPASSWORD "password" /w
    local db_name=${1:-postgres}
    shift
    psql -U "${SQL_USER:-postgres}" -h localhost -p $PORT "$@" $db_name
}

log_status_eq() {
    if ! [ "$status" -eq $1 ]; then
        echo "status: expected $1, received $status"
        printf "output:\n$output"
        exit 1
    fi
}

log_output_has() {
    if ! [[ "$output" =~ $1 ]]; then
        echo "output did not have $1"
        printf "output:\n$output"
        exit 1
    fi
}

nativevar DOLT_ROOT_PATH $BATS_TMPDIR/config-$$ /p
