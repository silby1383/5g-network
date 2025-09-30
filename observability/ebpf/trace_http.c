// SPDX-License-Identifier: GPL-2.0 OR BSD-3-Clause
/* Copyright (c) 2024 5G Network Project */

/*
 * NOTE: This eBPF program requires kernel headers to compile.
 * 
 * To install kernel headers:
 *   Ubuntu/Debian: sudo apt-get install linux-headers-$(uname -r)
 *   Fedora/RHEL:   sudo dnf install kernel-devel
 * 
 * The 5G network works without eBPF - this is for advanced observability.
 * See EBPF-SETUP.md for details.
 */

#ifdef __BPF__
/* This will only compile if proper headers are available */

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

/* Basic type definitions */
typedef unsigned char __u8;
typedef unsigned short __u16;
typedef unsigned int __u32;
typedef unsigned long long __u64;
typedef long long __s64;

#else
/* Placeholder when headers aren't available */
#warning "eBPF headers not available - creating placeholder"
typedef unsigned char __u8;
typedef unsigned short __u16;
typedef unsigned int __u32;
typedef unsigned long long __u64;
#endif

#define TRACEPARENT_LEN 55
#define MAX_PATH_LEN 128
#define MAX_METHOD_LEN 16

/* HTTP event structure */
struct http_event {
    __u64 timestamp_ns;
    __u32 pid;
    __u32 tid;
    char method[MAX_METHOD_LEN];
    char path[MAX_PATH_LEN];
    char traceparent[TRACEPARENT_LEN];
    __u16 status_code;
    __u64 duration_ns;
    __u32 content_length;
};

/* Map to export events to userspace */
struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(__u32));
} http_events SEC(".maps");

/* Map to store active requests (for calculating duration) */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, __u64);
    __type(value, struct http_event);
} active_requests SEC(".maps");

/* Map to store W3C trace context */
struct trace_context {
    __u8 trace_id[16];
    __u8 span_id[8];
    __u8 flags;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, __u64);
    __type(value, struct trace_context);
} trace_contexts SEC(".maps");

/* Helper function to parse traceparent header */
static __always_inline int parse_traceparent(const char *header, struct trace_context *ctx) {
    // Format: 00-{trace-id}-{span-id}-{flags}
    // Example: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
    
    if (!header || !ctx)
        return -1;

    // Simple parsing (in production, would need full implementation)
    // For now, just copy the header
    bpf_probe_read_kernel_str(&ctx->trace_id, sizeof(ctx->trace_id), header + 3);
    bpf_probe_read_kernel_str(&ctx->span_id, sizeof(ctx->span_id), header + 36);
    
    return 0;
}

/* Trace HTTP request start */
SEC("uprobe/http_handler_start")
int trace_http_request_start(struct pt_regs *ctx) {
    __u64 id = bpf_get_current_pid_tgid();
    __u64 ts = bpf_ktime_get_ns();
    
    struct http_event event = {};
    event.timestamp_ns = ts;
    event.pid = id >> 32;
    event.tid = id & 0xFFFFFFFF;
    
    // Read HTTP method from function parameter
    // Assuming first parameter is pointer to method string
    void *method_ptr = (void *)PT_REGS_PARM1(ctx);
    bpf_probe_read_user_str(&event.method, sizeof(event.method), method_ptr);
    
    // Read HTTP path from second parameter
    void *path_ptr = (void *)PT_REGS_PARM2(ctx);
    bpf_probe_read_user_str(&event.path, sizeof(event.path), path_ptr);
    
    // Read traceparent header from third parameter
    void *traceparent_ptr = (void *)PT_REGS_PARM3(ctx);
    if (traceparent_ptr) {
        bpf_probe_read_user_str(&event.traceparent, sizeof(event.traceparent), traceparent_ptr);
        
        // Parse and store trace context
        struct trace_context trace_ctx = {};
        parse_traceparent(event.traceparent, &trace_ctx);
        bpf_map_update_elem(&trace_contexts, &id, &trace_ctx, BPF_ANY);
    }
    
    // Store the event for later duration calculation
    bpf_map_update_elem(&active_requests, &id, &event, BPF_ANY);
    
    return 0;
}

/* Trace HTTP request end */
SEC("uprobe/http_handler_end")
int trace_http_request_end(struct pt_regs *ctx) {
    __u64 id = bpf_get_current_pid_tgid();
    __u64 ts = bpf_ktime_get_ns();
    
    // Lookup the stored event
    struct http_event *event = bpf_map_lookup_elem(&active_requests, &id);
    if (!event) {
        return 0;  // No matching request start
    }
    
    // Calculate duration
    event->duration_ns = ts - event->timestamp_ns;
    
    // Read status code from function parameter
    event->status_code = (__u16)PT_REGS_PARM1(ctx);
    
    // Read content length if available
    void *content_len_ptr = (void *)PT_REGS_PARM2(ctx);
    if (content_len_ptr) {
        bpf_probe_read_kernel(&event->content_length, sizeof(event->content_length), content_len_ptr);
    }
    
    // Send event to userspace
    bpf_perf_event_output(ctx, &http_events, BPF_F_CURRENT_CPU, event, sizeof(*event));
    
    // Cleanup
    bpf_map_delete_elem(&active_requests, &id);
    bpf_map_delete_elem(&trace_contexts, &id);
    
    return 0;
}

/* Trace function entry (generic uprobe) */
SEC("uprobe/function_entry")
int trace_function_entry(struct pt_regs *ctx) {
    __u64 id = bpf_get_current_pid_tgid();
    __u64 ts = bpf_ktime_get_ns();
    
    // Simple function tracing
    struct http_event event = {};
    event.timestamp_ns = ts;
    event.pid = id >> 32;
    event.tid = id & 0xFFFFFFFF;
    
    // Function name would be set by userspace based on which function we're tracing
    
    bpf_map_update_elem(&active_requests, &id, &event, BPF_ANY);
    
    return 0;
}

/* Trace function exit (generic uretprobe) */
SEC("uretprobe/function_exit")
int trace_function_exit(struct pt_regs *ctx) {
    __u64 id = bpf_get_current_pid_tgid();
    __u64 ts = bpf_ktime_get_ns();
    
    struct http_event *event = bpf_map_lookup_elem(&active_requests, &id);
    if (!event) {
        return 0;
    }
    
    event->duration_ns = ts - event->timestamp_ns;
    
    // Send to userspace
    bpf_perf_event_output(ctx, &http_events, BPF_F_CURRENT_CPU, event, sizeof(*event));
    
    bpf_map_delete_elem(&active_requests, &id);
    
    return 0;
}

/* Kprobe on TCP sendmsg for network tracing */
SEC("kprobe/tcp_sendmsg")
int trace_tcp_sendmsg(struct pt_regs *ctx) {
    __u64 ts = bpf_ktime_get_ns();
    __u64 id = bpf_get_current_pid_tgid();
    
    // Extract socket information
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    
    // Extract trace context if available
    struct trace_context *trace_ctx = bpf_map_lookup_elem(&trace_contexts, &id);
    if (trace_ctx) {
        // Trace context is available - this TCP send is part of a traced HTTP request
        // In production, would inject trace context into packet headers here
    }
    
    return 0;
}

/* Kprobe on TCP recvmsg for network tracing */
SEC("kprobe/tcp_recvmsg")
int trace_tcp_recvmsg(struct pt_regs *ctx) {
    __u64 ts = bpf_ktime_get_ns();
    
    // Extract socket information
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    
    // Would extract and parse TCP data for trace context
    
    return 0;
}

char _license[] SEC("license") = "Dual BSD/GPL";
