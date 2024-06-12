#pragma once

#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

typedef MaaBool (*RunCallback)(
                 MaaSyncContextHandle sync_context,
                 MaaStringView task_name,
                 MaaStringView custom_action_param,
                 MaaRectHandle cur_box,
                 MaaStringView cur_rec_detail,
                 MaaTransparentArg action_arg);

typedef void (*StopCallback)(MaaTransparentArg action_arg);

extern MaaCustomActionHandle MaaCustomActionHandleCreate(RunCallback run, StopCallback stop);

extern void MaaCustomActionHandleDestroy(MaaCustomActionHandle handle);