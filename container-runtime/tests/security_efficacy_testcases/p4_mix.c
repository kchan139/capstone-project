#include "fs_ops.h"
#include "stdio.h"
#include <time.h>
#define NUM_OPS 50

int main() {
    char filepath[64];
    int read_success = 0;
    int write_success = 0;
    int exec_success = 0;
    struct timespec start, end;

    clock_gettime(CLOCK_MONOTONIC, &start);

    // 50 reads
    for (int i = 1; i <= NUM_OPS; i++) {
        snprintf(filepath, sizeof(filepath), "test_4/read_%d.txt", i);
        if (do_open_read(filepath) == 0)
            read_success++;
    }

    // 50 writes
    for (int i = 1; i <= NUM_OPS; i++) {
        snprintf(filepath, sizeof(filepath), "test_4/write_%d.txt", i);
        if (do_open_write(filepath, "test content\n") == 0)
            write_success++;
    }

    // 50 execs
    for (int i = 1; i <= NUM_OPS; i++) {
        if (do_exec("test_4/dummy") == 0)
            exec_success++;
    }

    clock_gettime(CLOCK_MONOTONIC, &end);

    double elapsed = (end.tv_sec - start.tv_sec) +
                     (end.tv_nsec - start.tv_nsec) / 1e9;

    printf("[P4] reads: %d, writes: %d, execs: %d\n",
           read_success, write_success, exec_success);
    printf("Total time: %.6f seconds\n", elapsed);

    return 0;
}