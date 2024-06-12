#include "custom_recognizer.h"

MaaCustomRecognizerHandle MaaCustomRecognizerHandleCreate(AnalyzeCallback analyze) {
    MaaCustomRecognizerHandle handle = malloc(sizeof(struct MaaCustomRecognizerAPI));
    if (handle == NULL) {
            return NULL;
    }

    handle->analyze = analyze;
    return handle;
}

void MaaCustomRecognizerHandleDestroy(MaaCustomRecognizerHandle handle) {
    if (handle != NULL) {
            free(handle);
    };
}