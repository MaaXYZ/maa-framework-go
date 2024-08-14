#pragma once

#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

typedef MaaBool (*ConnectCallback)(MaaTransparentArg handle_arg);

        /// Write result to buffer.
typedef MaaBool (
            *RequestUUIDCallback)(MaaTransparentArg handle_arg, /* out */ MaaStringBufferHandle buffer);

typedef MaaBool (*StartAppCallback)(MaaStringView intent, MaaTransparentArg handle_arg);
typedef MaaBool (*StopAppCallback)(MaaStringView intent, MaaTransparentArg handle_arg);

        /// Write result to buffer.
typedef MaaBool (*ScreencapCallback)(MaaTransparentArg handle_arg, /* out */ MaaImageBufferHandle buffer);

typedef MaaBool (*ClickCallback)(int32_t x, int32_t y, MaaTransparentArg handle_arg);
typedef MaaBool (*SwipeCallback)(
            int32_t x1,
            int32_t y1,
            int32_t x2,
            int32_t y2,
            int32_t duration,
            MaaTransparentArg handle_arg);
typedef MaaBool (*TouchDownCallback)(
            int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            MaaTransparentArg handle_arg);
typedef MaaBool (*TouchMoveCallback)(
            int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            MaaTransparentArg handle_arg);
typedef MaaBool (*TouchUpCallback)(int32_t contact, MaaTransparentArg handle_arg);

typedef MaaBool (*PressKeyCallback)(int32_t keycode, MaaTransparentArg handle_arg);
typedef MaaBool (*InputTextCallback)(MaaStringView text, MaaTransparentArg handle_arg);

extern MaaCustomControllerHandle MaaCustomControllerHandleCreate(
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
);

extern void MaaCustomControllerHandleDestroy(MaaCustomControllerHandle handle);
