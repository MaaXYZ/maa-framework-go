#pragma once

#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

typedef MaaBool (*ConnectCallback)(void* handle_arg);

        /// Write result to buffer.
typedef MaaBool (
            *RequestUUIDCallback)(void*  handle_arg, /* out */ MaaStringBuffer* buffer);

typedef MaaBool (*StartAppCallback)(const char* intent, void*  handle_arg);
typedef MaaBool (*StopAppCallback)(const char* intent, void*  handle_arg);

        /// Write result to buffer.
typedef MaaBool (*ScreencapCallback)(void* handle_arg, /* out */ MaaImageBuffer* buffer);

typedef MaaBool (*ClickCallback)(int32_t x, int32_t y, void*  handle_arg);
typedef MaaBool (*SwipeCallback)(
            int32_t x1,
            int32_t y1,
            int32_t x2,
            int32_t y2,
            int32_t duration,
            void*  handle_arg);
typedef MaaBool (*TouchDownCallback)(
            int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            void*  handle_arg);
typedef MaaBool (*TouchMoveCallback)(
            int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            void*  handle_arg);
typedef MaaBool (*TouchUpCallback)(int32_t contact, void*  handle_arg);

typedef MaaBool (*PressKeyCallback)(int32_t keycode, void*  handle_arg);
typedef MaaBool (*InputTextCallback)(const char* text, void*  handle_arg);

extern MaaCustomControllerCallbacks* MaaCustomControllerHandleCreate(
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

extern void MaaCustomControllerHandleDestroy(MaaCustomControllerCallbacks* handle);
