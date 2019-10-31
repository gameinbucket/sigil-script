#ifndef HMLIB_SLICE_H
#define HMLIB_SLICE_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <stdarg.h>
#include <stdint.h>
#include <inttypes.h>

struct __attribute__((__packed__)) hmlib_slice_head
{
    size_t length;
    size_t capacity;
    void **slice;
};

typedef void *hmlib_slice;
typedef struct hmlib_slice_head hmlib_slice_head;

hmlib_slice hmlib_slice_init(const size_t member_size, const size_t length);
hmlib_slice hmlib_array_to_slice(void *const array, const size_t member_size, const size_t length);
void hmlib_slice_free(const hmlib_slice a);
size_t hmlib_slice_len(const hmlib_slice a);
int hmlib_slice_len_int(const hmlib_slice a);
hmlib_slice_head *hmlib_slice_resize(const hmlib_slice head, const size_t member_size, const size_t length);
hmlib_slice hmlib_slice_expand(const hmlib_slice a, const hmlib_slice b);
hmlib_slice hmlib_slice_push(const hmlib_slice a, void *const b);
hmlib_slice hmlib_slice_push_int(const hmlib_slice a, const int b);
hmlib_slice hmlib_slice_push_float(const hmlib_slice a, const float b);

#endif
