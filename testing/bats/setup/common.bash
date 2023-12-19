load setup/windows-compat
load setup/query-server-common

if [ -z "$BATS_TMPDIR" ]; then
    export BATS_TMPDIR=$HOME/batstmp/
    mkdir $BATS_TMPDIR
fi

nativebatsdir() { echo `nativepath $BATS_TEST_DIRNAME/$1`; }
batshelper() { echo `nativebatsdir helper/$1`; }

stash_current_dolt_user() {
    export STASHED_DOLT_USER_NAME=`doltgresql config --global --get user.name`
    export STASHED_DOLT_USER_EMAIL=`doltgresql config --global --get user.email`
}

restore_stashed_dolt_user() {
    doltgresql config --global --add user.name "$STASHED_DOLT_USER_NAME"
    doltgresql config --global --add user.email "$STASHED_DOLT_USER_EMAIL"
    unset STASHED_DOLT_USER_NAME STASHED_DOLT_USER_EMAIL
}

set_dolt_user() {
    doltgresql config --global --add user.name "$1" > /dev/null 2>&1
    doltgresql config --global --add user.email "$2" > /dev/null 2>&1
}

unset_dolt_user() {
  doltgresql config --global --unset user.name
  doltgresql config --global --unset user.email
}

current_dolt_user_name() {
    doltgresql config --global --get user.name
}

current_dolt_user_email() {
    doltgresql config --global --get user.email
}

setup_common() {
    run psql --version
    if [[ ! "$output" =~ "(PostgreSQL) 15" ]] && [[ ! "$output" =~ "(PostgreSQL) 16" ]]; then
        echo "PSQL must be version 15"
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
    nativevar DOLTGRES_DATA_DIR "$(pwd)" /p # This has to be set in every function that calls doltgresql
    nativevar DOLTGRES_DATA_DIR_CWD "" /w

    mkdir "postgres"
    cd "postgres"
    doltgresql init
    cd ..

    if [ -z "$DOLT_TEST_RETRIES" ]; then
        export BATS_TEST_RETRIES="$DOLT_TEST_RETRIES"
    fi

    start_sql_server
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
    psql -U "${SQL_USER:-postgres}" -h localhost -p $PORT "$@"
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
doltgresql config --global --add metrics.disabled true > /dev/null 2>&1
set_dolt_user "Bats Tests" "bats@email.fake" 
