#include <stdio.h>
#include "fs_ops.h"

#define NUM_EXEC 50

int main() {
    int success = 0;

    for (int i = 1; i <= NUM_EXEC; i++) {
        if (do_exec("./dummy") == 0)
            success++;
    }

    printf("\n[P3] Done. %d/%d executions successful.\n",
           success, NUM_EXEC);
    printf("[P3] Expected events: %d FAN_OPEN_EXEC\n",
           success);
    return 0;
}