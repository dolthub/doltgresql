/******************************************************************************
 *
 * Project:  PROJ
 * Purpose:  projinfo C API
 * Author:   Javier Jimenez Shaw
 *
 ******************************************************************************
 * Copyright (c) 2025  Javier Jimenez Shaw
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
 * OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 ****************************************************************************/

#if !defined(PROJAPPS_LIB_H)
#define PROJAPPS_LIB_H

#include "proj.h"

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Level for the output given by projinfo in its callback.
 */
typedef enum {
    PJ_PROJINFO_LOG_LEVEL_INFO = 1,
    PJ_PROJINFO_LOG_LEVEL_WARN = 2,
    PJ_PROJINFO_LOG_LEVEL_ERR = 3,
} PJ_PROJINFO_LOG_LEVEL;

typedef void (*projinfo_cb_t)(PJ_PROJINFO_LOG_LEVEL level, const char *msg,
                              void *user_data);

/*
 * Internal C implementation of projinfo CLI application.
 * See https://proj.org/apps/projinfo.html for more documentation.
 * @param ctx context. It can be nullptr.
 * @param argc number for parameters in argv.
 * @param argv list of char* with the command parameters of projinfo.
 *  It does not contain the program name as first parameter.
 * @param cb callback that to get the output of projinfo.
 *  It can be very fragmented, no necessarily by lines.
 * @param user_data pointer for data passed to the callback.
 */
int PROJ_DLL projinfo(PJ_CONTEXT *ctx, int argc, char **argv, projinfo_cb_t cb,
                      void *user_data);

#ifdef __cplusplus
}
#endif

#endif
