#include "custom_action.h"

MaaCustomActionHandle MaaCustomActionHandleCreate(RunCallback run, StopCallback stop) {
    MaaCustomActionHandle handle = malloc(sizeof(struct MaaCustomActionAPI));
    if (handle == NULL) {
        return NULL;
    }

    handle->run = run;
    handle->stop = stop;
    return handle;
}

void MaaCustomActionHandleDestroy(MaaCustomActionHandle handle) {
    if (handle != NULL) {
        free(handle);
    }
}