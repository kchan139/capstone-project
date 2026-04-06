#include <stdio.h>
#include "fs_ops.h"

#define NUM_FILES 50

int main() {
    char filepath[64];
    int success = 0;

    for (int i = 1; i <= NUM_FILES; i++) {
        snprintf(filepath, sizeof(filepath), "test_%d.txt", i);
        if (do_open_write(filepath, "test content\n") == 0)
            success++;
    }

    printf("\n[P2] Done. %d/%d files written successfully.\n",
           success, NUM_FILES);
    printf("[P2] Expected events: %d FAN_OPEN + %d FAN_MODIFY = %d total\n",
           success, success, success * 2);
    return 0;
}