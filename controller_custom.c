#include "controller_custom.h"

MaaCustomControllerHandle MaaCustomControllerHandleCreate(
    ConnectCallback connect,
    RequestUUIDCallback request_uuid,
    StartAppCallback start_app,
    StopAppCallback stop_app,
    ScreencapCallback screencap,
    ClickCallback click,
    SwipeCallback swipe,
    TouchDownCallback touch_down,
    TouchMoveCallback touch_move,
    TouchUpCallback touch_up,
    PressKeyCallback press_key,
    InputTextCallback input_text
) {
    MaaCustomControllerHandle handle = malloc(sizeof(struct MaaCustomControllerAPI));
    if (handle == NULL) {
        return NULL;
    }

    handle->connect = connect;
    handle->request_uuid = request_uuid;
    handle->start_app = start_app;
    handle->stop_app = stop_app;
    handle->screencap = screencap;
    handle->click = click;
    handle->swipe = swipe;
    handle->touch_down = touch_down;
    handle->touch_move = touch_move;
    handle->touch_up = touch_up;
    handle->press_key = press_key;
    handle->input_text = input_text;
    return handle;
}

void MaaCustomControllerHandleDestroy(MaaCustomControllerHandle handle) {
    if (handle != NULL) {
        free(handle);
    }
}