#pragma once

#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

typedef MaaBool (*AnalyzeCallback)(
            MaaSyncContextHandle sync_context,
            const MaaImageBufferHandle image,
            MaaStringView task_name,
            MaaStringView custom_recognition_param,
            MaaTransparentArg recognizer_arg,
            /*out*/ MaaRectHandle out_box,
            /*out*/ MaaStringBufferHandle out_detail);

extern MaaCustomRecognizerHandle MaaCustomRecognizerHandleCreate(AnalyzeCallback analyze);

extern void MaaCustomRecognizerHandleDestroy(MaaCustomRecognizerHandle handle);