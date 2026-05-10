#include <stdio.h>
#include "fs_ops.h"

#define NUM_EXEC 50

int main() {
    int success = 0;

    for (int i = 1; i <= NUM_EXEC; i++) {
        if (do_exec("test_3/dummy") == 0)
            success++;
    }
    return 0;
}