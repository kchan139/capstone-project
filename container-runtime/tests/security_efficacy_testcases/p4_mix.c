#include "fs_ops.h"
#include "stdio.h"
#define NUM_OPS 20

int main() {
    char filepath[64];
    int read_success = 0;
    int write_success = 0;
    int exec_success = 0;

    // 20 reads
    for (int i = 1; i <= NUM_OPS; i++) {
        snprintf(filepath, sizeof(filepath), "read_%d.txt", i);
        if (do_open_read(filepath) == 0)
            read_success++;
    }

    // 20 writes
    for (int i = 1; i <= NUM_OPS; i++) {
        snprintf(filepath, sizeof(filepath), "write_%d.txt", i);
        if (do_open_write(filepath, "test content\n") == 0)
            write_success++;
    }

    // 20 execs
    for (int i = 1; i <= NUM_OPS; i++) {
        if (do_exec("./dummy") == 0)
            exec_success++;
    }

    return 0;
}