#ifndef X11_H
#define X11_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

char* multihash_x11(const char* input, uint32_t len);

#ifdef __cplusplus
}
#endif

#endif
