#include "hmlib_slice.h"

hmlib_slice_head *hmlib_slice_head_init(const size_t member_size, const size_t length, const size_t capacity) {
    size_t memory = sizeof(hmlib_slice_head) + capacity * member_size;
    hmlib_slice_head *head = hmlib_malloc(memory);
    memset(head, 0, memory);
    head->length = length;
    head->capacity = capacity;
    return head;
}

hmlib_slice_head *hmlib_slice_get_head(const hmlib_slice a) {
    return (hmlib_slice_head *)a - 1;
}

hmlib_slice hmlib_slice_init(const size_t member_size, const size_t length, const size_t capacity) {
    hmlib_slice_head *head = hmlib_slice_head_init(member_size, length, capacity);
    return (hmlib_slice_head *)head + 1;
}

hmlib_slice hmlib_slice_simple_init(const size_t member_size, const size_t length) {
    return hmlib_slice_init(member_size, length, length);
}

hmlib_slice hmlib_array_to_slice(void *const array, const size_t member_size, const size_t length) {
    size_t array_memory = length * member_size;
    size_t memory = sizeof(hmlib_slice_head) + array_memory;
    hmlib_slice_head *new_head = hmlib_malloc(memory);
    new_head->length = length;
    new_head->capacity = length;
    hmlib_slice_head *data = (hmlib_slice_head *)new_head + 1;
    memcpy(data, array, array_memory);
    return data;
}

void hmlib_slice_free(const hmlib_slice a) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    free(head);
}

size_t hmlib_slice_len_size(const hmlib_slice a) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    return head->length;
}

int hmlib_slice_len(const hmlib_slice a) {
    return (int)hmlib_slice_len_size(a);
}

size_t hmlib_slice_cap_size(const hmlib_slice a) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    return head->capacity;
}

int hmlib_slice_cap(const hmlib_slice a) {
    return (int)hmlib_slice_cap_size(a);
}

hmlib_slice_head *hmlib_slice_resize(const hmlib_slice head, const size_t member_size, const size_t length) {
    size_t memory = sizeof(hmlib_slice_head) + length * member_size;
    hmlib_slice_head *new_head = hmlib_realloc(head, memory);
    new_head->length = length;
    new_head->capacity = length;
    return new_head;
}

hmlib_slice hmlib_slice_expand(const hmlib_slice a, const hmlib_slice b) {
    hmlib_slice_head *head_a = hmlib_slice_get_head(a);
    hmlib_slice_head *head_b = hmlib_slice_get_head(b);
    size_t length_a = head_a->length;
    size_t length_b = head_b->length;
    size_t length = length_a + length_b;
    hmlib_slice_head *new_head = hmlib_slice_resize(head_a, sizeof(void *), length);
    hmlib_slice_head *data = (hmlib_slice_head *)new_head + 1;
    memcpy(data + length_a * sizeof(void *), b, length_b * sizeof(void *));
    return data;
}

hmlib_slice hmlib_slice_push(const hmlib_slice a, void *const b) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    size_t length = head->length + 1;
    if (length > head->capacity) {
        head = hmlib_slice_resize(head, sizeof(void *), length);
        head->capacity = length;
    }
    head->length = length;
    hmlib_slice data = (hmlib_slice_head *)head + 1;
    ((hmlib_slice_head **)data)[length - 1] = b;
    return data;
}

hmlib_slice hmlib_slice_push_int(const hmlib_slice a, const int b) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    size_t length = head->length + 1;
    if (length > head->capacity) {
        head = hmlib_slice_resize(head, sizeof(b), length);
        head->capacity = length;
    }
    head->length = length;
    hmlib_slice data = (hmlib_slice_head *)head + 1;
    ((int *)data)[length - 1] = b;
    return data;
}

hmlib_slice hmlib_slice_push_float(const hmlib_slice a, const float b) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    size_t length = head->length + 1;
    if (length > head->capacity) {
        head = hmlib_slice_resize(head, sizeof(b), length);
        head->capacity = length;
    }
    head->length = length;
    hmlib_slice data = (hmlib_slice_head *)head + 1;
    ((float *)data)[length - 1] = b;
    return data;
}

void *hmlib_slice_pop(const hmlib_slice a) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    size_t length = head->length;
    if (length == 0) {
        return 0;
    }
    head->length--;
    hmlib_slice data = (hmlib_slice_head *)head + 1;
    return ((hmlib_slice_head **)data)[length - 1];
}

int hmlib_slice_pop_int(const hmlib_slice a) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    size_t length = head->length;
    if (length == 0) {
        return 0;
    }
    head->length--;
    hmlib_slice data = (hmlib_slice_head *)head + 1;
    return ((int *)data)[length - 1];
}

float hmlib_slice_pop_float(const hmlib_slice a) {
    hmlib_slice_head *head = hmlib_slice_get_head(a);
    size_t length = head->length;
    if (length == 0) {
        return 0;
    }
    head->length--;
    hmlib_slice data = (hmlib_slice_head *)head + 1;
    return ((float *)data)[length - 1];
}
