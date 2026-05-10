#include <stdio.h>
#include <time.h>
#include "fs_ops.h"

#define NUM_EXEC 50

int main() {
    int success = 0;
    struct timespec start, end;

    clock_gettime(CLOCK_MONOTONIC, &start);

    for (int i = 1; i <= NUM_EXEC; i++) {
        if (do_exec("test_3/dummy") == 0)
            success++;
    }

    clock_gettime(CLOCK_MONOTONIC, &end);

    double elapsed = (end.tv_sec - start.tv_sec) +
                     (end.tv_nsec - start.tv_nsec) / 1e9;

    printf("Total time: %.6f seconds\n", elapsed);

    return 0;
}