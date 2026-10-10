/* Amalgamated source file */
#include "upb/generated_code_support.h"

/*
 * This is where we define internal portability macros used across upb.
 *
 * All of these macros are undef'd in undef.inc to avoid leaking them to users.
 *
 * The correct usage is:
 *
 *   #include "upb/foobar.h"
 *   #include "upb/baz.h"
 *
 *   // MUST be last included header.
 *   #include "upb/port/def.inc"
 *
 *   // Code for this file.
 *   // <...>
 *
 *   // Can be omitted for .c files, required for .h.
 *   #include "upb/port/undef.inc"
 *
 * This file is private and must not be included by users!
 */

#if !((defined(__STDC_VERSION__) && __STDC_VERSION__ >= 199901L) || \
      (defined(__cplusplus) && __cplusplus >= 201703L) ||           \
      (defined(_MSC_VER) && _MSC_VER >= 1900))
#error upb requires C99 or C++17 or MSVC >= 2015.
#endif

// Portable check for GCC minimum version:
// https://gcc.gnu.org/onlinedocs/cpp/Common-Predefined-Macros.html
#if defined(__GNUC__) && defined(__GNUC_MINOR__) && defined(__GNUC_PATCHLEVEL__)
#define UPB_GNUC_MIN(x, y) \
  (__GNUC__ > (x) || __GNUC__ == (x) && __GNUC_MINOR__ >= (y))
#else
#define UPB_GNUC_MIN(x, y) 0
#endif

#if defined(__clang__) && defined(__clang_major__) && defined(__clang_minor__)
#define UPB_CLANG_MIN(x, y) \
  (__clang_major__ > (x) || __clang_major__ == (x) && __clang_minor__ >= (y))
#else
#define UPB_CLANG_MIN(x, y) 0
#endif

// Macros for checking for compiler attributes, defined here to avoid the
// problem described in
// https://gcc.gnu.org/onlinedocs/cpp/_005f_005fhas_005fattribute.html.
#ifdef __has_attribute
#define UPB_HAS_ATTRIBUTE(x) __has_attribute(x)
#else
#define UPB_HAS_ATTRIBUTE(x) 0
#endif

#ifdef __has_c_attribute
#define UPB_HAS_C_ATTRIBUTE(x) __has_c_attribute(x)
#else
#define UPB_HAS_C_ATTRIBUTE(x) 0
#endif

#if defined(__cplusplus) && defined(__has_cpp_attribute)
// NOTE: requiring __cplusplus above should not be necessary, but
// works around https://bugs.llvm.org/show_bug.cgi?id=23435.
#define UPB_HAS_CPP_ATTRIBUTE(x) __has_cpp_attribute(x)
#else
#define UPB_HAS_CPP_ATTRIBUTE(x) 0
#endif

// Once in a while we want to use this macro in a C++-only portion of an
// otherwise C-compatible header, so we copy and paste this from ABSL.
#if UPB_HAS_CPP_ATTRIBUTE(deprecated) && UPB_HAS_CPP_ATTRIBUTE(clang::annotate)
#define UPB_DEPRECATE_AND_INLINE() [[deprecated, clang::annotate("inline-me")]]
#elif UPB_HAS_CPP_ATTRIBUTE(deprecated)
#define UPB_DEPRECATE_AND_INLINE() [[deprecated]]
#else
#define UPB_DEPRECATE_AND_INLINE()
#endif

#ifdef __has_builtin
#define UPB_HAS_BUILTIN(x) __has_builtin(x)
#else
#define UPB_HAS_BUILTIN(x) 0
#endif

#ifdef __has_extension
#define UPB_HAS_EXTENSION(x) __has_extension(x)
#else
#define UPB_HAS_EXTENSION(x) 0
#endif

#ifdef __has_feature
#define UPB_HAS_FEATURE(x) __has_feature(x)
#else
#define UPB_HAS_FEATURE(x) 0
#endif

#include <assert.h>
#include <setjmp.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#ifndef UINTPTR_MAX
Error, UINTPTR_MAX is undefined
#endif

#if UINTPTR_MAX == 0xffffffff
#define UPB_SIZE(size32, size64) size32
#else
#define UPB_SIZE(size32, size64) size64
#endif

/* If we always read/write as a consistent type to each address, this shouldn't
 * violate aliasing.
 */
#define UPB_PTR_AT(msg, ofs, type) ((type*)((char*)(msg) + (ofs)))

// A flexible array member may have lower alignment requirements than the struct
// overall - in that case, it can overlap with the trailing padding of the rest
// of the struct, and a naive sizeof(base) + sizeof(flex) * count calculation
// will not take into account that overlap, and allocate more than is required.
#define UPB_SIZEOF_FLEX(type, member, count) \
  UPB_MAX(sizeof(type), offsetof(type, member[count]))

#define UPB_SIZEOF_FLEX_WOULD_OVERFLOW(type, member, count) \
  (((SIZE_MAX - offsetof(type, member[0])) /                \
    (offsetof(type, member[1]) - offsetof(type, member[0]))) < (size_t)count)

#define UPB_ARRAY_SIZE(arr) (sizeof(arr) / sizeof(arr[0]))

#define UPB_MAPTYPE_STRING 0

// UPB_EXPORT: always generate a public symbol.
#if defined(__GNUC__) || defined(__clang__)
#define UPB_EXPORT __attribute__((visibility("default"))) __attribute__((used))
#else
#define UPB_EXPORT
#endif

// UPB_INLINE: inline if possible, emit standalone code if required.
#ifdef __cplusplus
#define UPB_INLINE inline
#elif defined(__GNUC__) || defined(__clang__)
#define UPB_INLINE static __inline__
#else
#define UPB_INLINE static
#endif

// UPB_INLINE_IF_NOT_GCC: because gcc can be very noisy at times.
#if defined(__GNUC__) && !defined(__clang__)
#define UPB_INLINE_IF_NOT_GCC static
#else
#define UPB_INLINE_IF_NOT_GCC UPB_INLINE
#endif

#ifdef UPB_BUILD_API
#define UPB_API UPB_EXPORT
#define UPB_API_INLINE UPB_EXPORT
#else
#define UPB_API
#define UPB_API_INLINE UPB_INLINE
#endif

#ifdef EXPORT_UPBC
#define UPBC_API UPB_EXPORT
#else
#define UPBC_API
#endif

#if UPB_HAS_FEATURE(address_sanitizer) || defined(__SANITIZE_ADDRESS__)
#define UPB_ASAN 1
#else
#define UPB_ASAN 0
#endif

#if UPB_HAS_FEATURE(hwaddress_sanitizer)
#define UPB_HWASAN 1
#define UPB_HWASAN_POISON_TAG 17
#define UPB_MALLOC_ALIGN 16
#else
#define UPB_HWASAN 0
#define UPB_MALLOC_ALIGN 8
#endif

#if UPB_HAS_FEATURE(thread_sanitizer) || defined(__SANITIZE_THREAD__)
#define UPB_TSAN 1
#else
#define UPB_TSAN 0
#endif

#if UPB_HAS_FEATURE(memory_sanitizer)
#define UPB_MSAN 1
#else
#define UPB_MSAN 0
#endif

// An unfortunate concession to C++17 and MSVC, which don't support zero-sized
// structs.
#if UPB_ASAN || UPB_HWASAN || UPB_TSAN
#define UPB_XSAN_MEMBER upb_Xsan xsan;
#define UPB_XSAN(st) (&(st)->xsan)
#define UPB_XSAN_STRUCT_SIZE 1
#else
#define UPB_XSAN_MEMBER
#define UPB_XSAN(st) (NULL)
#define UPB_XSAN_STRUCT_SIZE 0
#endif

#define UPB_ALIGN_UP(size, align) (((size) + (align) - 1) / (align) * (align))
#define UPB_ALIGN_DOWN(size, align) ((size) / (align) * (align))
#define UPB_ALIGN_MALLOC(size) UPB_ALIGN_UP(size, UPB_MALLOC_ALIGN)

#if __STDC_VERSION__ >= 202311L || UPB_HAS_EXTENSION(cxx_alignof) || \
    defined(__cplusplus)
#define UPB_ALIGN_OF(type) alignof(type)
#elif __STDC_VERSION__ >= 201112L || UPB_HAS_EXTENSION(c_alignof)
#define UPB_ALIGN_OF(type) _Alignof(type)
#elif UPB_GNUC_MIN(2, 95)
#define UPB_ALIGN_OF(type) __alignof__(type)
#elif defined(_MSC_VER)
#define UPB_ALIGN_OF(type) __alignof(type)
#else
#define UPB_ALIGN_OF(type) \
  offsetof(                \
      struct {             \
        char c;            \
        type member;       \
      },                   \
      member)
#endif

#ifdef _MSC_VER
// Some versions of our Windows compiler don't support the C11 syntax.
#define UPB_ALIGN_AS(x) __declspec(align(x))
#elif defined(__GNUC__)
#define UPB_ALIGN_AS(x) __attribute__((aligned(x)))
#else
#define UPB_ALIGN_AS(x) _Alignas(x)
#endif

#if __STDC_VERSION__ >= 202311L || UPB_HAS_EXTENSION(cxx_static_assert) || \
    defined(__cplusplus)
#define UPB_STATIC_ASSERT(val, msg) static_assert((val), msg)
#elif __STDC_VERSION__ >= 201112L || UPB_HAS_EXTENSION(c_static_assert) || \
    UPB_GNUC_MIN(4, 6)
#define UPB_STATIC_ASSERT(val, msg) _Static_assert((val), msg)
#else
// Unfortunately this hack doesn't work inside struct declarations, but it works
// everywhere else
#define UPB_STATIC_ASSERT_CONCAT_IMPL(s1, s2) s1##s2
#define UPB_STATIC_ASSERT_CONCAT(s1, s2) UPB_STATIC_ASSERT_CONCAT_IMPL(s1, s2)
#ifdef __COUNTER__
#define UPB_STATIC_ASSERT(condition, message)                      \
  typedef char UPB_STATIC_ASSERT_CONCAT(static_assertion_failure_, \
                                        __COUNTER__)[(condition) ? 1 : -1]
#else
#define UPB_STATIC_ASSERT(condition, message)                      \
  typedef char UPB_STATIC_ASSERT_CONCAT(static_assertion_failure_, \
                                        __LINE__)[(condition) ? 1 : -1]
#endif
#endif

// Hints to the compiler about likely/unlikely branches.
#if defined(__GNUC__) || defined(__clang__)
#define UPB_LIKELY(x) __builtin_expect((bool)(x), 1)
#define UPB_UNLIKELY(x) __builtin_expect((bool)(x), 0)
#else
#define UPB_LIKELY(x) (x)
#define UPB_UNLIKELY(x) (x)
#endif

#if UPB_HAS_BUILTIN(__builtin_expect_with_probability)
#define UPB_UNPREDICTABLE(x) \
  __builtin_expect_with_probability((bool)(x), 1, 0.5)
#else
#define UPB_UNPREDICTABLE(x) (x)
#endif

// Macros for function attributes on compilers that support them.
#if defined(__GNUC__) || defined(__clang__)
#define UPB_FORCEINLINE __inline__ __attribute__((always_inline)) static
#define UPB_NOINLINE __attribute__((noinline))
#define UPB_NORETURN __attribute__((__noreturn__))
#define UPB_PRINTF(str, first_vararg) \
  __attribute__((format(printf, str, first_vararg)))
#elif defined(_MSC_VER)
#define UPB_NOINLINE
#define UPB_FORCEINLINE static
#define UPB_NORETURN __declspec(noreturn)
#define UPB_PRINTF(str, first_vararg)
#else /* !defined(__GNUC__) */
#define UPB_FORCEINLINE static
#define UPB_NOINLINE
#define UPB_NORETURN
#define UPB_PRINTF(str, first_vararg)
#endif

#if UPB_HAS_ATTRIBUTE(noderef)
#define UPB_NODEREF __attribute__((noderef))
#else
#define UPB_NODEREF
#endif

// Will be defined properly once call sites are updated
#if false && UPB_HAS_C_ATTRIBUTE(nodiscard)
#define UPB_NODISCARD [[nodiscard]]
#elif false && UPB_HAS_ATTRIBUTE(warn_unused_result)
#define UPB_NODISCARD __attribute__((warn_unused_result))
#elif false && UPB_HAS_CPP_ATTRIBUTE(nodiscard)
#define UPB_NODISCARD [[nodiscard]]
#else
#define UPB_NODISCARD
#endif

#define UPB_MAX(x, y) ((x) > (y) ? (x) : (y))
#define UPB_MIN(x, y) ((x) < (y) ? (x) : (y))

#define UPB_UNUSED(var) (void)(var)

// UPB_ASSUME(): in release mode, we tell the compiler to assume this is true.
#ifdef NDEBUG
#ifdef __GNUC__
#define UPB_ASSUME(expr) \
  if (!(expr)) __builtin_unreachable()
#elif defined _MSC_VER
#define UPB_ASSUME(expr) \
  if (!(expr)) __assume(0)
#else
#define UPB_ASSUME(expr) \
  do {                   \
  } while (false && (expr))
#endif
#else
#define UPB_ASSUME(expr) assert(expr)
#endif

#if UPB_HAS_BUILTIN(__builtin_constant_p) && UPB_HAS_ATTRIBUTE(const)
#define UPB_MAYBE_ASSUME(pred, x) \
  if (__builtin_constant_p(pred) && pred) UPB_ASSUME(x)
#define UPB_ATTR_CONST __attribute__((const))
#else
#define UPB_MAYBE_ASSUME(pred, x)
#define UPB_ATTR_CONST
#endif

/* UPB_ASSERT(): in release mode, we use the expression without letting it be
 * evaluated.  This prevents "unused variable" warnings. */
#ifdef NDEBUG
#define UPB_ASSERT(expr) \
  do {                   \
  } while (false && (expr))
#else
#define UPB_ASSERT(expr) assert(expr)
#endif

#if !defined(NDEBUG) && !defined(UPB_TSAN)
#define UPB_ENABLE_REF_CYCLE_CHECKS 1
#else
#define UPB_ENABLE_REF_CYCLE_CHECKS 0
#endif

#if defined(__GNUC__) || defined(__clang__)
#define UPB_PRETTY_FUNCTION __PRETTY_FUNCTION__
#else
#define UPB_PRETTY_FUNCTION NULL
#endif

#ifdef __cplusplus
    extern "C" {
#endif
  UPB_NORETURN void _upb_UnreachableFailure(const char* file, int line,
                                            const char* function_name);
#if !defined(NDEBUG)
#define UPB_UNREACHABLE_FAILURE() \
  _upb_UnreachableFailure(__FILE__, __LINE__, UPB_PRETTY_FUNCTION);
#else
#define UPB_UNREACHABLE_FAILURE()
#endif
#ifdef __cplusplus
} /* extern "C" */
#endif

#if defined(__GNUC__) || defined(__clang__)
#define UPB_UNREACHABLE()      \
  do {                         \
    UPB_UNREACHABLE_FAILURE(); \
    __builtin_unreachable();   \
  } while (0)
#elif defined(_MSC_VER)
#define UPB_UNREACHABLE()      \
  do {                         \
    UPB_UNREACHABLE_FAILURE(); \
    __assume(0);               \
  } while (0)
#else
#define UPB_UNREACHABLE()      \
  do {                         \
    UPB_UNREACHABLE_FAILURE(); \
  } while (0)
#endif

#ifdef __ANDROID__
#define UPB_DEFAULT_MAX_BLOCK_SIZE 8192
#else
#define UPB_DEFAULT_MAX_BLOCK_SIZE 32768
#endif

/* UPB_SETJMP() / UPB_LONGJMP() */
// Android uses a custom libc that does not implement all of posix, but it has
// had sigsetjmp/siglongjmp forever on arm and since API 12 on x86. Apple has
// sigsetjmp, but does not define the posix feature test macro.
#if defined(__APPLE__) || defined(_POSIX_C_SOURCE) || defined(__ANDROID__)
// avoid setting/restoring signal mask, which involves costly syscalls
#define UPB_SETJMP(buf) sigsetjmp(buf, 0)
#define UPB_LONGJMP(buf, val) siglongjmp(buf, val)
#elif defined(WASM_WAMR)
#define UPB_SETJMP(buf) 0
#define UPB_LONGJMP(buf, val) abort()
#else
#define UPB_SETJMP(buf) setjmp(buf)
#define UPB_LONGJMP(buf, val) longjmp(buf, val)
#endif

#if ((__STDC_VERSION__ >= 201112L) && !defined(__STDC_NO_ATOMICS__)) || \
    UPB_HAS_EXTENSION(c_atomic) ||                                      \
    defined(__GNUC__)  // GCC supported atomics as an extension before it
                       // supported __has_extension
#define UPB_USE_C11_ATOMICS
#elif defined(_MSC_VER)
#define UPB_USE_MSC_ATOMICS
#endif

#if defined(UPB_USE_C11_ATOMICS)
#define UPB_ATOMIC(T) _Atomic(T)
#elif defined(UPB_USE_MSC_ATOMICS)
#define UPB_ATOMIC(T) volatile T
#else
#define UPB_ATOMIC(T) T
#endif

/* UPB_PTRADD(ptr, ofs): add pointer while avoiding "NULL + 0" UB */
#define UPB_PTRADD(ptr, ofs) ((ofs) ? (ptr) + (ofs) : (ptr))

#define UPB_PRIVATE(x) x##_dont_copy_me__upb_internal_use_only

#ifdef UPB_ALLOW_PRIVATE_ACCESS__FOR_BITS_ONLY
#define UPB_ONLYBITS(x) x
#else
#define UPB_ONLYBITS(x) UPB_PRIVATE(x)
#endif

/* Configure whether fasttable is switched on or not. *************************/

#if UPB_HAS_ATTRIBUTE(musttail)
#define UPB_MUSTTAIL __attribute__((musttail))
#else
#define UPB_MUSTTAIL
#endif

#if UPB_HAS_ATTRIBUTE(preserve_most) && !defined(__i386__)
#define UPB_PRESERVE_MOST __attribute__((preserve_most))
#else
#define UPB_PRESERVE_MOST
#endif

#if UPB_HAS_ATTRIBUTE(preserve_none)
#define UPB_PRESERVE_NONE __attribute__((preserve_none))
#else
#define UPB_PRESERVE_NONE
#endif

#if defined(__aarch64__) && (defined(__GNUC__) || defined(__clang__))
#define UPB_ARM64_ASM 1
#else
#define UPB_ARM64_ASM 0
#endif

/* When compiling with branch protection, we need to ensure that all branch
 * targets in assembly use the appropriate landing pad instruction. These
 * instructions are backwards compatible with processors that don't have
 * FEAT_BTI and are treated as nops.
 */
#if UPB_ARM64_ASM && defined(__ARM_FEATURE_BTI_DEFAULT)
#if __ARM_FEATURE_BTI_DEFAULT == 1
#define UPB_ARM64_BTI_DEFAULT 1
#else
#define UPB_ARM64_BTI_DEFAULT 0
#endif
#else
#define UPB_ARM64_BTI_DEFAULT 0
#endif

/* aarch64 supports big and little endian modes; fasttable performs multibyte
 * tag loads assumes the tag of a varint is in the low bits. */
#if (defined(__x86_64__) || defined(__AARCH64EL__)) && \
    UPB_HAS_ATTRIBUTE(preserve_none) && UPB_HAS_ATTRIBUTE(musttail)
#define UPB_FASTTABLE_SUPPORTED 1
#else
#define UPB_FASTTABLE_SUPPORTED 0
#endif

/* define UPB_ENABLE_FASTTABLE to force fast table support.
 * This is useful when we want to ensure we are really getting fasttable,
 * for example for testing or benchmarking. */
#if defined(UPB_ENABLE_FASTTABLE)
#if !UPB_FASTTABLE_SUPPORTED
#error fasttable is x86-64/ARM64le only and requires preserve_none and musttail.
#endif
#define UPB_FASTTABLE 1
/* Define UPB_TRY_ENABLE_FASTTABLE to use fasttable if possible.
 * This is useful for releasing code that might be used on multiple platforms,
 * for example the PHP or Ruby C extensions. */
#elif defined(UPB_TRY_ENABLE_FASTTABLE)
#define UPB_FASTTABLE UPB_FASTTABLE_SUPPORTED
#else
#define UPB_FASTTABLE 0
#endif

/* UPB_FASTTABLE_INIT() allows protos compiled for fasttable to gracefully
 * degrade to non-fasttable if the runtime or platform do not support it. */
#if !UPB_FASTTABLE
#define UPB_FASTTABLE_INIT(...)
#define UPB_FASTTABLE_MASK(mask) -1
#else
#define UPB_FASTTABLE_INIT(...) __VA_ARGS__
#define UPB_FASTTABLE_MASK(mask) mask
#endif

#undef UPB_FASTTABLE_SUPPORTED

#if defined(__cplusplus)
#if defined(__clang__) || UPB_GNUC_MIN(6, 0)
// https://gcc.gnu.org/gcc-6/changes.html
#define UPB_DEPRECATED [[deprecated]]
#else
#define UPB_DEPRECATED
#endif
#else
#define UPB_DEPRECATED
#endif

#if defined(UPB_IS_GOOGLE3) && \
    (!defined(UPB_BOOTSTRAP_STAGE) || UPB_BOOTSTRAP_STAGE != 0)
#define UPB_DESC_MINITABLE(sym) &proto2__##sym##_msg_init
#elif defined(UPB_IS_GOOGLE3) && defined(UPB_BOOTSTRAP_STAGE) && \
    UPB_BOOTSTRAP_STAGE == 0
#define UPB_DESC_MINITABLE(sym) proto2__##sym##_msg_init()
#elif defined(UPB_BOOTSTRAP_STAGE) && UPB_BOOTSTRAP_STAGE == 0
#define UPB_DESC_MINITABLE(sym) google__protobuf__##sym##_msg_init()
#else
#define UPB_DESC_MINITABLE(sym) &google__protobuf__##sym##_msg_init
#endif

#undef UPB_IS_GOOGLE3

#ifdef __clang__
#define UPB_NO_SANITIZE_ADDRESS __attribute__((no_sanitize("address")))
#else
#define UPB_NO_SANITIZE_ADDRESS
#endif

#if defined(__GNUC__) && (defined(__clang__) || UPB_GNUC_MIN(11, 0))
#define UPB_RETAIN __attribute__((retain))
#else
#define UPB_RETAIN
#endif

// Linker arrays combine elements from multiple translation units into a single
// array that can be iterated over at runtime.
//
// It is an alternative to pre-main "registration" functions.
//
// Usage:
//
//   // In N translation units.
//   UPB_LINKARR_APPEND(foo_array) static int elems[3] = {1, 2, 3};
//
//   // At runtime:
//   UPB_LINKARR_DECLARE(foo_array, int);
//
//   void f() {
//     const int* start = UPB_LINKARR_START(foo_array);
//     const int* stop = UPB_LINKARR_STOP(foo_array);
//     for (const int* p = start; p < stop; p++) {
//       // Windows can introduce zero padding, so we have to skip zeroes.
//       if (*p != 0) {
//         vec.push_back(*p);
//       }
//     }
//   }

#if defined(__GNUC__) && !defined(__clang__)
// GCC can't handle mismatched retain attributes in the same section:
//   https://github.com/protocolbuffers/protobuf/issues/26385
// To work around this, we retain all linker array elements, even though this
// effectively disables tree-shaking of unused extensions when using GCC.
#define UPB_LINKARR_ATTR UPB_RETAIN
#else
#define UPB_LINKARR_ATTR
#endif

#define UPB_LINKARR_SENTINEL UPB_RETAIN __attribute__((weak, used))

#if defined(__ELF__) || defined(__wasm__)

#define UPB_LINKARR_APPEND(name) \
  __attribute__((                \
      section("linkarr_" #name))) UPB_LINKARR_ATTR UPB_NO_SANITIZE_ADDRESS
#define UPB_LINKARR_DECLARE(name, type)                                       \
  UPB_STATIC_ASSERT(sizeof("__la_" #name) <= 17,                              \
                    "Linker array name too long for Mach-O (16-char limit)"); \
  extern type __start_linkarr_##name;                                         \
  extern type __stop_linkarr_##name;                                          \
  UPB_LINKARR_APPEND(name)                                                    \
  UPB_LINKARR_SENTINEL type UPB_linkarr_internal_empty_##name[1]
#define UPB_LINKARR_START(name) (&__start_linkarr_##name)
#define UPB_LINKARR_STOP(name) (&__stop_linkarr_##name)

#elif defined(__MACH__)

/* As described in: https://stackoverflow.com/a/22366882 */
#define UPB_LINKARR_APPEND(name) \
  __attribute__((                \
      section("__DATA,__la_" #name))) UPB_LINKARR_ATTR UPB_NO_SANITIZE_ADDRESS
#define UPB_LINKARR_DECLARE(name, type)                                       \
  UPB_STATIC_ASSERT(sizeof("__la_" #name) <= 17,                              \
                    "Linker array name too long for Mach-O (16-char limit)"); \
  extern type __start_linkarr_##name __asm(                                   \
      "section$start$__DATA$__la_" #name);                                    \
  extern type __stop_linkarr_##name __asm(                                    \
      "section$end$__DATA$"                                                   \
      "__la_" #name);                                                         \
  UPB_LINKARR_APPEND(name)                                                    \
  UPB_LINKARR_SENTINEL type UPB_linkarr_internal_empty_##name[1]
#define UPB_LINKARR_START(name) (&__start_linkarr_##name)
#define UPB_LINKARR_STOP(name) (&__stop_linkarr_##name)

#elif defined(_MSC_VER)

/* See:
 *   https://devblogs.microsoft.com/oldnewthing/20181107-00/?p=100155
 *   https://devblogs.microsoft.com/oldnewthing/20181108-00/?p=100165
 *   https://devblogs.microsoft.com/oldnewthing/20181109-00/?p=100175 */
#define UPB_STRINGIFY_INTERNAL(x) #x
#define UPB_STRINGIFY(x) UPB_STRINGIFY_INTERNAL(x)
#define UPB_CONCAT(a, b, c) a##b##c
#define UPB_LINKARR_NAME(name, index) \
  UPB_STRINGIFY(UPB_CONCAT(la_, name, index))
#define UPB_LINKARR_APPEND(name)                      \
  __pragma(section(UPB_LINKARR_NAME(name, $j), read)) \
      __declspec(allocate(UPB_LINKARR_NAME(name, $j)))
// clang-format off
#define UPB_LINKARR_DECLARE(name, type)                          \
  __pragma(message(UPB_LINKARR_NAME(name, $j)))                  \
  __pragma(section(UPB_LINKARR_NAME(name, $a), read))            \
  __pragma(section(UPB_LINKARR_NAME(name, $z), read))            \
  __declspec(allocate(UPB_LINKARR_NAME(name, $a)), selectany)    \
            type __start_linkarr_##name;                         \
  __declspec(allocate(UPB_LINKARR_NAME(name, $z)), selectany)    \
            type __stop_linkarr_##name;                          \
  UPB_LINKARR_APPEND(name)                                       \
  __declspec(selectany) type UPB_linkarr_internal_empty_##name[1] = {0}
// clang-format on
#define UPB_LINKARR_START(name) (&__start_linkarr_##name)
#define UPB_LINKARR_STOP(name) (&__stop_linkarr_##name)

#else

// Linker arrays are not supported on this platform.  Make macros no-ops.
#define UPB_LINKARR_APPEND(name)
#define UPB_LINKARR_DECLARE(name, type)          \
  UPB_STATIC_ASSERT(sizeof("__la_" #name) <= 17, \
                    "Linker array name too long for Mach-O (16-char limit)")
#define UPB_LINKARR_START(name) (NULL)
#define UPB_LINKARR_STOP(name) (NULL)

#endif

// Workaround for https://github.com/llvm/llvm-project/issues/167577 until it's
// fixed.  Some function must exist for the constructor to work properly.
// TODO Remove this or gate it on a future version of clang.
#if defined(__clang__) && defined(__arm__)
#define _UPB_CONSTRUCTOR_PLACEHOLDER(unique_name)                            \
  __attribute__((used, visibility("hidden"))) void UPB_PRIVATE(unique_name)( \
      void) {}
#else
#define _UPB_CONSTRUCTOR_PLACEHOLDER(unique_name)
#endif

#define _UPB_STRINGIFY2(x) #x
#define _UPB_STRINGIFY(x) _UPB_STRINGIFY2(x)

#if defined(__ELF__) && (UPB_GNUC_MIN(15, 1) || UPB_CLANG_MIN(21, 1))
/*
 * Workaround for b/456308964 and b/456317163. Although weak constructors
 * resolve to a single function body, Clang still emits a pointer into
 * .init_array for every translation unit, causing the constructor to be
 * executed multiple times and inflating binary size. We wrap the .init_array
 * entry in a COMDAT group (using inline assembly) to force the linker to
 * deduplicate the pointer to exactly one instance. %cc is used in preference to
 * the more widely available %c because %c may under some circumstances still
 * output a function name including relocation information, which will then
 * confuse the linker. This was only a problem on gcc; when %cc was introduced
 * to clang it was just an alias to %c, and clang's %c had already done the
 * checks that %cc introduced to gcc.
 * References:
 * https://github.com/gcc-mirror/gcc/commit/74d6a676034b3ab20c387f12f19f5597e4f1c9fa
 * https://github.com/llvm/llvm-project/pull/127719#issuecomment-2686276305
 */
#define UPB_CONSTRUCTOR(name, unique_name, ...)                             \
  _UPB_CONSTRUCTOR_PLACEHOLDER(unique_name)                                 \
  __attribute__((weak, used, visibility("hidden"))) void UPB_PRIVATE(name)( \
      void) {                                                               \
    __asm__ volatile(                                                       \
        ".pushsection .init_array,\"awG\",%%init_array, %cc0, comdat\n"     \
        ".dc.a %cc0\n"                                                      \
        ".popsection\n"                                                     \
        :                                                                   \
        : "X"(UPB_PRIVATE(name)));                                          \
    __VA_ARGS__                                                             \
  }
#elif defined(__ELF__) || defined(__wasm__) || defined(__MACH__) || \
    defined(__MINGW32__)
#define UPB_CONSTRUCTOR(name, unique_name, ...)                              \
  _UPB_CONSTRUCTOR_PLACEHOLDER(unique_name)                                  \
  __attribute__((weak, visibility("hidden"), constructor)) void UPB_PRIVATE( \
      name)(void) {                                                          \
    __VA_ARGS__                                                              \
  }
#elif defined(_MSC_VER)
/*
 * See: https://stackoverflow.com/questions/1113409
 *
 * The /include pragma suggested in the link above doesn't work in our case
 * because it requires globally unique names. We need a different solution
 * to prevent optimizers from removing the constructor. Our solution is to
 * create a dummy exported weak symbol that prevent this stripping.
 */
#pragma section(".CRT$XCT", read)
#define UPB_CONSTRUCTOR(name, unique_name, ...)                              \
  static void __cdecl UPB_PRIVATE(name)(void);                               \
  __declspec(allocate(".CRT$XCT"), selectany) void(                          \
      __cdecl * UPB_PRIVATE(name##_))(void) = UPB_PRIVATE(name);             \
  __declspec(selectany, dllexport) void* UPB_PRIVATE(name##_force_linkage) = \
      &UPB_PRIVATE(name##_);                                                 \
  static void __cdecl UPB_PRIVATE(name)(void) { __VA_ARGS__ }
#else
// No constructor support, nothing we can do except not break builds.
#if defined(__GNUC__) || defined(__clang__)
#define UPB_CONSTRUCTOR(name, unique_name, ...) \
  static __attribute__((used)) void UPB_PRIVATE(name)(void) { __VA_ARGS__ }
#else
#define UPB_CONSTRUCTOR(name, unique_name, ...) \
  static void UPB_PRIVATE(name)(void) { __VA_ARGS__ }
#endif
#endif

//
// Weak alias platform support. Theoretically this should be possible to do with
// only C using attributes like __attribute__((weak, alias("foo"))), but
// Clang doesn't support this properly on macOS.  So we have to use assembly.
#if defined(__APPLE__)

// TODO: once https://github.com/llvm/llvm-project/issues/167262 is fixed
// in the LLVM linker, we should have all weak variables point to a single
// "default" empty MiniTable instead of having each leaf define its own, like
// we do with ELF below. This will reduce binary size if many messages are tree
// shaken.
#define UPB_WEAK_SINGLETON_PLACEHOLDER_MINITABLE()
#define UPB_WEAK_PLACEHOLDER_MINITABLE(name)           \
  __attribute__((weak)) const upb_MiniTable name = {   \
      .UPB_PRIVATE(fields) = NULL,                     \
      .UPB_PRIVATE(size) = sizeof(struct upb_Message), \
      .UPB_PRIVATE(field_count) = 0,                   \
      .UPB_PRIVATE(ext) = kUpb_ExtMode_NonExtendable,  \
      .UPB_PRIVATE(dense_below) = 0,                   \
      .UPB_PRIVATE(table_mask) = -1,                   \
      .UPB_PRIVATE(required_count) = 0,                \
  };
#define UPB_WEAK_ALIAS(type, from, to) \
  extern type to;                      \
  __asm__(".globl _" #to);             \
  __asm__(".private_extern _" #to);    \
  __asm__(".set _" #to ", _" #from);   \
  __asm__(".weak_definition _" #to);
#define UPB_STRONG_ALIAS(type, from, to) \
  __asm__(".globl _" #to);               \
  __asm__(".private_extern _" #to);      \
  __asm__(".set _" #to ", _" #from);

#elif defined(__ELF__) || defined(__wasm__)

  // On ELF, weak aliases work properly, so we can have all weak MiniTables
  // point to the same empty singleton MiniTable. This reduces code size if many
  // MiniTables are tree shaken.
#define UPB_WEAK_SINGLETON_PLACEHOLDER_MINITABLE()               \
  __attribute__((weak))                                          \
  const upb_MiniTable kUpb_WeakSingletonPlaceholderMiniTable = { \
      .UPB_PRIVATE(fields) = NULL,                               \
      .UPB_PRIVATE(size) = sizeof(struct upb_Message),           \
      .UPB_PRIVATE(field_count) = 0,                             \
      .UPB_PRIVATE(ext) = kUpb_ExtMode_NonExtendable,            \
      .UPB_PRIVATE(dense_below) = 0,                             \
      .UPB_PRIVATE(table_mask) = -1,                             \
      .UPB_PRIVATE(required_count) = 0,                          \
  };
#define UPB_WEAK_PLACEHOLDER_MINITABLE(name)
#define UPB_WEAK_ALIAS(type, from, to) \
  extern type to                       \
      __attribute__((weak, alias("kUpb_WeakSingletonPlaceholderMiniTable")));
#define UPB_STRONG_ALIAS(type, from, to) \
  extern type to __attribute__((alias(#from)));

#else
#define UPB_WEAK_SINGLETON_PLACEHOLDER_MINITABLE()
#define UPB_WEAK_PLACEHOLDER_MINITABLE(name)
#define UPB_WEAK_ALIAS(type, from, to) weak_alias_not_supported_on_this_platform
#define UPB_STRONG_ALIAS(type, from, to) \
  strong_alias_not_supported_on_this_platform

#endif

#ifndef UPB_GENERATED_CODE_SUPPORT_H_
#define UPB_GENERATED_CODE_SUPPORT_H_

// This is a bit awkward; we want to conditionally include the fast decoder,
// but we generally don't let macros like UPB_FASTTABLE leak into user code.
// We can't #include "decode_fast.h" inside the port/def.inc, because the
// inc files strictly prohibit recursive inclusion, and decode_fast.h includes
// port/def.inc. So instead we use this two-part dance to conditionally include
// decode_fast.h.
#if UPB_FASTTABLE
#define UPB_INCLUDE_FAST_DECODE
#endif

// IWYU pragma: begin_exports
#ifdef UPB_INCLUDE_FAST_DECODE
#endif
// IWYU pragma: end_exports

#undef UPB_INCLUDE_FAST_DECODE

#endif  // UPB_GENERATED_CODE_SUPPORT_H_


#include <errno.h>
#include <float.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>

// Must be last.

void upb_Status_Clear(upb_Status* status) {
  if (!status) return;
  status->ok = true;
  status->msg[0] = '\0';
}

bool upb_Status_IsOk(const upb_Status* status) { return status->ok; }

const char* upb_Status_ErrorMessage(const upb_Status* status) {
  return status->msg;
}

void upb_Status_SetErrorMessage(upb_Status* status, const char* msg) {
  if (!status) return;
  status->ok = false;
  strncpy(status->msg, msg, _kUpb_Status_MaxMessage - 1);
  status->msg[_kUpb_Status_MaxMessage - 1] = '\0';
}

void upb_Status_SetErrorFormat(upb_Status* status, const char* fmt, ...) {
  va_list args;
  va_start(args, fmt);
  upb_Status_VSetErrorFormat(status, fmt, args);
  va_end(args);
}

void upb_Status_VSetErrorFormat(upb_Status* status, const char* fmt,
                                va_list args) {
  if (!status) return;
  status->ok = false;
  vsnprintf(status->msg, sizeof(status->msg), fmt, args);
  status->msg[_kUpb_Status_MaxMessage - 1] = '\0';
}

void upb_Status_VAppendErrorFormat(upb_Status* status, const char* fmt,
                                   va_list args) {
  size_t len;
  if (!status) return;
  status->ok = false;
  len = strlen(status->msg);
  vsnprintf(status->msg + len, sizeof(status->msg) - len, fmt, args);
  status->msg[_kUpb_Status_MaxMessage - 1] = '\0';
}

/*
 * upb_table Implementation
 *
 * Implementation is heavily inspired by Lua's ltable.c.
 */


#include <stdint.h>
#include <string.h>


// Must be last.

#if defined(__has_builtin)
#if __has_builtin(__builtin_popcount)
#define UPB_FAST_POPCOUNT32(i) __builtin_popcount(i)
#endif
#elif defined(__GNUC__)
#define UPB_FAST_POPCOUNT32(i) __builtin_popcount(i)
#elif defined(_MSC_VER) && (defined(_M_IX86) || defined(_M_X64))
// Only use __popcnt on x86/x64 architectures for MSVC
#define UPB_FAST_POPCOUNT32(i) __popcnt(i)
#endif

UPB_INLINE int _upb_popcnt32(uint32_t i) {
#ifdef UPB_FAST_POPCOUNT32
  return UPB_FAST_POPCOUNT32(i);
#else
  int count = 0;
  while (i != 0) {
    count += i & 1;
    i >>= 1;
  }
  return count;
#endif
}

#undef UPB_FAST_POPCOUNT32

UPB_INLINE uint8_t _upb_log2_table_size(upb_table* t) {
  return _upb_popcnt32(t->mask);
}

/* A type to represent the lookup key of either a strtable, inttable or
 * exttable. */
typedef union {
  uintptr_t num;
  upb_StringView str;
  struct {
    const void* ptr;
    uint32_t ext_num;
  } ext;
} lookupkey_t;

static lookupkey_t strkey2(const char* str, size_t len) {
  return (lookupkey_t){.str = upb_StringView_FromDataAndSize(str, len)};
}

static lookupkey_t intkey(uintptr_t key) { return (lookupkey_t){.num = key}; }

static lookupkey_t extkey(const void* ptr, uint32_t ext_num) {
  return (lookupkey_t){.ext = {ptr, ext_num}};
}

// Conceptually the hash and equal functions should only take the key, not the
// value, but the extension table stores part of its logical key in the value
// slot. This is a sign that we have outgrown the original architecture.
typedef uint32_t hashfunc_t(upb_key key, upb_value val);
typedef bool eqlfunc_t(upb_key k1, upb_value v1, lookupkey_t k2);

/* Base table (shared code) ***************************************************/

static uint32_t upb_inthash(uintptr_t key) {
  UPB_STATIC_ASSERT(sizeof(uintptr_t) == 4 || sizeof(uintptr_t) == 8,
                    "Pointers don't fit");
  if (sizeof(uintptr_t) == 8) {
    return (uint32_t)key ^ (uint32_t)((uint64_t)key >> 32);
  } else {
    return (uint32_t)key;
  }
}

static upb_tabent* upb_getentry(const upb_table* t, uint32_t hash) {
  return t->entries + (hash & t->mask);
}

static bool isfull(upb_table* t) {
  uint32_t size = upb_table_size(t);
  // 0.875 load factor
  return t->count == (size - (size >> 3));
}

static bool init(upb_table* t, uint8_t size_lg2, upb_Arena* a) {
  if (size_lg2 >= 32) {
    return false;
  }
  t->count = 0;
  uint32_t size = 1U << size_lg2;
  t->mask = size - 1;  // 0 mask if size_lg2 is 0
  if (upb_table_size(t) > (SIZE_MAX / sizeof(upb_tabent))) {
    return false;
  }
  size_t bytes = upb_table_size(t) * sizeof(upb_tabent);
  if (bytes > 0) {
    t->entries = upb_Arena_Malloc(a, bytes);
    if (!t->entries) return false;
    memset(t->entries, 0, bytes);
  } else {
    t->entries = NULL;
  }
  return true;
}

static upb_tabent* emptyent(upb_table* t, upb_tabent* e) {
  upb_tabent* begin = t->entries;
  upb_tabent* end = begin + upb_table_size(t);
  for (e = e + 1; e < end; e++) {
    if (upb_tabent_isempty(e)) return e;
  }
  for (e = begin; e < end; e++) {
    if (upb_tabent_isempty(e)) return e;
  }
  UPB_ASSERT(false);
  return NULL;
}

static upb_tabent* getentry_mutable(upb_table* t, uint32_t hash) {
  return upb_getentry(t, hash);
}

static upb_tabent* findentry(const upb_table* t, lookupkey_t key, uint32_t hash,
                             eqlfunc_t* eql) {
  upb_tabent* e;

  if (t->count == 0) return NULL;
  e = upb_getentry(t, hash);
  if (upb_tabent_isempty(e)) return NULL;
  while (1) {
    if (eql(e->key, e->val, key)) return e;
    if (!upb_tabent_hasnext(e)) return NULL;
    e = upb_tabent_next(e);
  }
}

static upb_tabent* findentry_mutable(upb_table* t, lookupkey_t key,
                                     uint32_t hash, eqlfunc_t* eql) {
  return findentry(t, key, hash, eql);
}

static bool lookup(const upb_table* t, lookupkey_t key, upb_value* v,
                   uint32_t hash, eqlfunc_t* eql) {
  const upb_tabent* e = findentry(t, key, hash, eql);
  if (e) {
    if (v) *v = e->val;
    return true;
  } else {
    return false;
  }
}

/* The given key must not already exist in the table. */
static void insert(upb_table* t, lookupkey_t key, upb_key tabkey, upb_value val,
                   uint32_t hash, hashfunc_t* hashfunc, eqlfunc_t* eql) {
  upb_tabent* mainpos_e;
  upb_tabent* our_e;

  UPB_ASSERT(findentry(t, key, hash, eql) == NULL);

  t->count++;
  mainpos_e = getentry_mutable(t, hash);
  our_e = mainpos_e;

  if (upb_tabent_isempty(mainpos_e)) {
    /* Our main position is empty; use it. */
    upb_tabent_clearnext(our_e);
  } else {
    /* Collision. */
    upb_tabent* new_e = emptyent(t, mainpos_e);
    /* Head of collider's chain. */
    upb_tabent* chain =
        getentry_mutable(t, hashfunc(mainpos_e->key, mainpos_e->val));
    if (chain == mainpos_e) {
      /* Existing ent is in its main position (it has the same hash as us, and
       * is the head of our chain).  Insert to new ent and append to this chain.
       */
      if (upb_tabent_hasnext(mainpos_e)) {
        upb_tabent_setnext(new_e, upb_tabent_next(mainpos_e));
      } else {
        upb_tabent_clearnext(new_e);
      }
      upb_tabent_setnext(mainpos_e, new_e);
      our_e = new_e;
    } else {
      /* Existing ent is not in its main position (it is a node in some other
       * chain).  This implies that no existing ent in the table has our hash.
       * Evict it (updating its chain) and use its ent for head of our chain. */
      new_e->key = mainpos_e->key;
      new_e->val = mainpos_e->val;
      if (upb_tabent_hasnext(mainpos_e)) {
        upb_tabent_setnext(new_e, upb_tabent_next(mainpos_e));
      } else {
        upb_tabent_clearnext(new_e);
      }

      while (upb_tabent_hasnext(chain) && upb_tabent_next(chain) != mainpos_e) {
        chain = upb_tabent_next(chain);
        UPB_ASSERT(chain);
      }
      upb_tabent_setnext(chain, new_e);
      our_e = mainpos_e;
      upb_tabent_clearnext(our_e);
    }
  }
  our_e->key = tabkey;
  our_e->val = val;
  UPB_ASSERT(findentry(t, key, hash, eql) == our_e);
}

static bool rm(upb_table* t, lookupkey_t key, upb_value* val, uint32_t hash,
               eqlfunc_t* eql) {
  upb_tabent* chain = getentry_mutable(t, hash);
  if (upb_tabent_isempty(chain)) return false;
  if (eql(chain->key, chain->val, key)) {
    /* Element to remove is at the head of its chain. */
    t->count--;
    if (val) *val = chain->val;
    if (upb_tabent_hasnext(chain)) {
      upb_tabent* move = upb_tabent_next(chain);
      chain->key = move->key;
      chain->val = move->val;
      if (upb_tabent_hasnext(move)) {
        upb_tabent_setnext(chain, upb_tabent_next(move));
      } else {
        upb_tabent_clearnext(chain);
      }

      upb_tabent_clear(move);
    } else {
      upb_tabent_clear(chain);
    }
    return true;
  } else {
    /* Element to remove is either in a non-head position or not in the
     * table. */
    while (
        upb_tabent_hasnext(chain) &&
        !eql(upb_tabent_next(chain)->key, upb_tabent_next(chain)->val, key)) {
      chain = upb_tabent_next(chain);
    }
    if (upb_tabent_hasnext(chain)) {
      /* Found element to remove. */
      upb_tabent* rm = upb_tabent_next(chain);
      t->count--;
      if (val) *val = rm->val;
      if (upb_tabent_hasnext(rm)) {
        upb_tabent_setnext(chain, upb_tabent_next(rm));
      } else {
        upb_tabent_clearnext(chain);
      }
      upb_tabent_clear(rm);
      return true;
    } else {
      /* Element to remove is not in the table. */
      return false;
    }
  }
}

static size_t next(const upb_table* t, size_t i) {
  do {
    if (++i >= upb_table_size(t)) return SIZE_MAX - 1; /* Distinct from -1. */
  } while (upb_tabent_isempty(&t->entries[i]));

  return i;
}

static size_t begin(const upb_table* t) { return next(t, -1); }

UPB_FORCEINLINE
bool _upb_tablenext(const upb_table* t, upb_tabent** ent, intptr_t* iter) {
  size_t tab_idx = next(t, *iter);
  if (tab_idx < upb_table_size(t)) {
    *ent = &t->entries[tab_idx];
    *iter = tab_idx;
    return true;
  }
  return false;
}

UPB_FORCEINLINE
bool _upb_table_done(const upb_table* t, intptr_t iter) {
  if (iter == INTPTR_MAX - 1 || (size_t)iter >= upb_table_size(t)) {
    return true;
  }
  return upb_tabent_isempty(&t->entries[iter]);
}

static void removeiter(upb_table* t, intptr_t* iter) {
  intptr_t i = *iter;
  upb_tabent* ent = &t->entries[i];
  upb_tabent* prev = NULL;

  // Linear search, not great.
  upb_tabent* end = &t->entries[upb_table_size(t)];
  for (upb_tabent* e = t->entries; e != end; e++) {
    if (!upb_tabent_isempty(e) && upb_tabent_hasnext(e) &&
        upb_tabent_next(e) == ent) {
      prev = e;
      break;
    }
  }

  if (prev) {
    if (upb_tabent_hasnext(ent)) {
      upb_tabent_setnext(prev, upb_tabent_next(ent));
    } else {
      upb_tabent_clearnext(prev);
    }
  } else {
    // ent is the head of the chain, so we need to move its next element into
    // its slot if there is one.
    if (upb_tabent_hasnext(ent)) {
      upb_tabent* move = upb_tabent_next(ent);
      ent->key = move->key;
      ent->val = move->val;
      if (upb_tabent_hasnext(move)) {
        upb_tabent_setnext(ent, upb_tabent_next(move));
      } else {
        upb_tabent_clearnext(ent);
      }

      // If we moved an element from a higher index to a lower index, then we've
      // moved an element we haven't visited yet into the slot of the one that
      // was just removed; decrement iter so that the iterator visits it.
      if (move > ent) {
        *iter = i - 1;
      }

      ent = move;
    }
  }

  t->count--;
  upb_tabent_clear(ent);
}

/* upb_strtable ***************************************************************/

// A simple "subclass" of upb_table that only adds a hash function for strings.

static upb_SizePrefixString* upb_SizePrefixString_Copy(upb_StringView s,
                                                       upb_Arena* a) {
  // A 2GB string will fail at serialization time, but we accept up to 4GB in
  // memory here.
  if (s.size > UINT32_MAX) return NULL;
  upb_SizePrefixString* str =
      upb_Arena_Malloc(a, sizeof(uint32_t) + s.size + 1);
  if (str == NULL) return NULL;
  str->size = s.size;
  char* data = (char*)str->data;
  if (s.size) memcpy(data, s.data, s.size);
  data[s.size] = '\0';
  return str;
}

/* Adapted from ABSL's wyhash. */

static uint64_t UnalignedLoad64(const void* p) {
  uint64_t val;
  memcpy(&val, p, 8);
  return val;
}

static uint32_t UnalignedLoad32(const void* p) {
  uint32_t val;
  memcpy(&val, p, 4);
  return val;
}

#if defined(_MSC_VER) && defined(_M_X64)
#include <intrin.h>
#endif

/* Computes a * b, returning the low 64 bits of the result and storing the high
 * 64 bits in |*high|. */
static uint64_t upb_umul128(uint64_t v0, uint64_t v1, uint64_t* out_high) {
#ifdef __SIZEOF_INT128__
  __uint128_t p = v0;
  p *= v1;
  *out_high = (uint64_t)(p >> 64);
  return (uint64_t)p;
#elif defined(_MSC_VER) && defined(_M_X64)
  return _umul128(v0, v1, out_high);
#else
  uint64_t a32 = v0 >> 32;
  uint64_t a00 = v0 & 0xffffffff;
  uint64_t b32 = v1 >> 32;
  uint64_t b00 = v1 & 0xffffffff;
  uint64_t high = a32 * b32;
  uint64_t low = a00 * b00;
  uint64_t mid1 = a32 * b00;
  uint64_t mid2 = a00 * b32;
  low += (mid1 << 32) + (mid2 << 32);
  // Omit carry bit, for mixing we do not care about exact numerical precision.
  high += (mid1 >> 32) + (mid2 >> 32);
  *out_high = high;
  return low;
#endif
}

static uint64_t WyhashMix(uint64_t v0, uint64_t v1) {
  uint64_t high;
  uint64_t low = upb_umul128(v0, v1, &high);
  return low ^ high;
}

static uint64_t Wyhash(const void* data, size_t len, uint64_t seed,
                       const uint64_t salt[]) {
  const uint8_t* ptr = (const uint8_t*)data;
  uint64_t starting_length = (uint64_t)len;
  uint64_t current_state = seed ^ salt[0];

  if (len > 64) {
    // If we have more than 64 bytes, we're going to handle chunks of 64
    // bytes at a time. We're going to build up two separate hash states
    // which we will then hash together.
    uint64_t duplicated_state = current_state;

    do {
      uint64_t a = UnalignedLoad64(ptr);
      uint64_t b = UnalignedLoad64(ptr + 8);
      uint64_t c = UnalignedLoad64(ptr + 16);
      uint64_t d = UnalignedLoad64(ptr + 24);
      uint64_t e = UnalignedLoad64(ptr + 32);
      uint64_t f = UnalignedLoad64(ptr + 40);
      uint64_t g = UnalignedLoad64(ptr + 48);
      uint64_t h = UnalignedLoad64(ptr + 56);

      uint64_t cs0 = WyhashMix(a ^ salt[1], b ^ current_state);
      uint64_t cs1 = WyhashMix(c ^ salt[2], d ^ current_state);
      current_state = (cs0 ^ cs1);

      uint64_t ds0 = WyhashMix(e ^ salt[3], f ^ duplicated_state);
      uint64_t ds1 = WyhashMix(g ^ salt[4], h ^ duplicated_state);
      duplicated_state = (ds0 ^ ds1);

      ptr += 64;
      len -= 64;
    } while (len > 64);

    current_state = current_state ^ duplicated_state;
  }

  // We now have a data `ptr` with at most 64 bytes and the current state
  // of the hashing state machine stored in current_state.
  while (len > 16) {
    uint64_t a = UnalignedLoad64(ptr);
    uint64_t b = UnalignedLoad64(ptr + 8);

    current_state = WyhashMix(a ^ salt[1], b ^ current_state);

    ptr += 16;
    len -= 16;
  }

  // We now have a data `ptr` with at most 16 bytes.
  uint64_t a = 0;
  uint64_t b = 0;
  if (len > 8) {
    // When we have at least 9 and at most 16 bytes, set A to the first 64
    // bits of the input and B to the last 64 bits of the input. Yes, they will
    // overlap in the middle if we are working with less than the full 16
    // bytes.
    a = UnalignedLoad64(ptr);
    b = UnalignedLoad64(ptr + len - 8);
  } else if (len > 3) {
    // If we have at least 4 and at most 8 bytes, set A to the first 32
    // bits and B to the last 32 bits.
    a = UnalignedLoad32(ptr);
    b = UnalignedLoad32(ptr + len - 4);
  } else if (len > 0) {
    // If we have at least 1 and at most 3 bytes, read all of the provided
    // bits into A, with some adjustments.
    a = ((ptr[0] << 16) | (ptr[len >> 1] << 8) | ptr[len - 1]);
    b = 0;
  } else {
    a = 0;
    b = 0;
  }

  uint64_t w = WyhashMix(a ^ salt[1], b ^ current_state);
  uint64_t z = salt[1] ^ starting_length;
  return WyhashMix(w, z);
}

const uint64_t kWyhashSalt[5] = {
    0x243F6A8885A308D3ULL, 0x13198A2E03707344ULL, 0xA4093822299F31D0ULL,
    0x082EFA98EC4E6C89ULL, 0x452821E638D01377ULL,
};

uint32_t _upb_Hash(const void* p, size_t n, uint64_t seed) {
  return Wyhash(p, n, seed, kWyhashSalt);
}

static const void* const _upb_seed;

// Returns a random seed for upb's hash function. This does not provide
// high-quality randomness, but it should be enough to prevent unit tests from
// relying on a deterministic map ordering. By returning the address of a
// variable, we are able to get some randomness for free provided that ASLR is
// enabled.
static uint64_t _upb_Seed(void) { return (uint64_t)&_upb_seed; }

static uint32_t _upb_Hash_NoSeed(const char* p, size_t n) {
  return _upb_Hash(p, n, _upb_Seed());
}

static uint32_t strhash(upb_key key, upb_value val) {
  UPB_UNUSED(val);
  return _upb_Hash_NoSeed(key.str->data, key.str->size);
}

static bool streql(upb_key k1, upb_value v1, lookupkey_t k2) {
  UPB_UNUSED(v1);
  const upb_SizePrefixString* k1s = k1.str;
  const upb_StringView k2s = k2.str;
  return k1s->size == k2s.size &&
         (k1s->size == 0 || memcmp(k1s->data, k2s.data, k1s->size) == 0);
}

/** Calculates the number of entries required to hold an expected number of
 * values, within the table's load factor. */
static size_t _upb_entries_needed_for(size_t expected_size) {
  size_t need_entries = expected_size + 1 + expected_size / 7;
  UPB_ASSERT(need_entries - (need_entries >> 3) >= expected_size);
  return need_entries;
}

bool upb_strtable_init(upb_strtable* t, size_t expected_size, upb_Arena* a) {
  int size_lg2 = upb_Log2Ceiling(_upb_entries_needed_for(expected_size));
  return init(&t->t, size_lg2, a);
}

void upb_strtable_clear(upb_strtable* t) {
  size_t bytes = upb_table_size(&t->t) * sizeof(upb_tabent);
  t->t.count = 0;
  memset((char*)t->t.entries, 0, bytes);
}

bool upb_strtable_resize(upb_strtable* t, size_t size_lg2, upb_Arena* a) {
  upb_strtable new_table;
  if (!init(&new_table.t, size_lg2, a)) return false;

  intptr_t iter = UPB_STRTABLE_BEGIN;
  upb_StringView sv;
  upb_value val;
  while (upb_strtable_next2(t, &sv, &val, &iter)) {
    // Unlike normal insert, does not copy string data or possibly reallocate
    // the table
    // The data pointer used in the table is guaranteed to point at a
    // upb_SizePrefixString, we just need to back up by the size of the uint32_t
    // length prefix.
    const upb_SizePrefixString* keystr =
        (const upb_SizePrefixString*)(sv.data - sizeof(uint32_t));
    UPB_ASSERT(keystr->data == sv.data);
    UPB_ASSERT(keystr->size == sv.size);

    lookupkey_t lookupkey = {.str = sv};
    upb_key tabkey = {.str = keystr};
    uint32_t hash = _upb_Hash_NoSeed(sv.data, sv.size);
    insert(&new_table.t, lookupkey, tabkey, val, hash, &strhash, &streql);
  }
  *t = new_table;
  return true;
}

bool upb_strtable_insert(upb_strtable* t, const char* k, size_t len,
                         upb_value v, upb_Arena* a) {
  if (isfull(&t->t)) {
    /* Need to resize.  New table of double the size, add old elements to it. */
    if (!upb_strtable_resize(t, _upb_log2_table_size(&t->t) + 1, a)) {
      return false;
    }
  }

  upb_StringView sv = upb_StringView_FromDataAndSize(k, len);
  upb_SizePrefixString* size_prefix_string = upb_SizePrefixString_Copy(sv, a);
  if (!size_prefix_string) return false;

  lookupkey_t lookupkey = {.str = sv};
  upb_key key = {.str = size_prefix_string};
  uint32_t hash = _upb_Hash_NoSeed(k, len);
  insert(&t->t, lookupkey, key, v, hash, &strhash, &streql);
  return true;
}

bool upb_strtable_lookup2(const upb_strtable* t, const char* key, size_t len,
                          upb_value* v) {
  uint32_t hash = _upb_Hash_NoSeed(key, len);
  return lookup(&t->t, strkey2(key, len), v, hash, &streql);
}

bool upb_strtable_remove2(upb_strtable* t, const char* key, size_t len,
                          upb_value* val) {
  uint32_t hash = _upb_Hash_NoSeed(key, len);
  return rm(&t->t, strkey2(key, len), val, hash, &streql);
}

/* Iteration */

void upb_strtable_begin(upb_strtable_iter* i, const upb_strtable* t) {
  i->t = t;
  i->index = begin(&t->t);
}

void upb_strtable_next(upb_strtable_iter* i) {
  i->index = next(&i->t->t, i->index);
}

bool upb_strtable_done(const upb_strtable_iter* i) {
  if (!i->t) return true;
  return _upb_table_done(&i->t->t, i->index);
}

upb_StringView upb_strtable_iter_key(const upb_strtable_iter* i) {
  UPB_ASSERT(!upb_strtable_done(i));
  return upb_key_strview(str_tabent(i)->key);
}

upb_value upb_strtable_iter_value(const upb_strtable_iter* i) {
  UPB_ASSERT(!upb_strtable_done(i));
  return str_tabent(i)->val;
}

void upb_strtable_iter_setdone(upb_strtable_iter* i) {
  i->t = NULL;
  i->index = SIZE_MAX;
}

bool upb_strtable_iter_isequal(const upb_strtable_iter* i1,
                               const upb_strtable_iter* i2) {
  if (upb_strtable_done(i1) && upb_strtable_done(i2)) return true;
  return i1->t == i2->t && i1->index == i2->index;
}

bool upb_strtable_next2(const upb_strtable* t, upb_StringView* key,
                        upb_value* val, intptr_t* iter) {
  upb_tabent* ent;
  if (_upb_tablenext(&t->t, &ent, iter)) {
    *key = upb_key_strview(ent->key);
    *val = ent->val;
    return true;
  }
  return false;
}

void upb_strtable_removeiter(upb_strtable* t, intptr_t* iter) {
  removeiter(&t->t, iter);
}

void upb_strtable_setentryvalue(upb_strtable* t, intptr_t iter, upb_value v) {
  t->t.entries[iter].val = v;
}

/* upb_exttable ***************************************************************/

static uint32_t _upb_exttable_hash(const void* ptr, uint32_t ext_num) {
  uint64_t a = (uintptr_t)ptr;
  uint64_t b = ext_num;
  return (uint32_t)WyhashMix(a ^ kWyhashSalt[1], b ^ _upb_Seed());
}

static uint32_t exthash(upb_key key, upb_value val) {
  const void* ptr = (const void*)key.num;
  uint32_t ext_num = *(const uint32_t*)upb_value_getconstptr(val);
  return _upb_exttable_hash(ptr, ext_num);
}

static bool exteql(upb_key k1, upb_value v1, lookupkey_t k2) {
  if ((const void*)k1.num == k2.ext.ptr) {
    uint32_t ext_num1 = *(const uint32_t*)upb_value_getconstptr(v1);
    return ext_num1 == k2.ext.ext_num;
  }
  return false;
}

bool upb_exttable_init(upb_exttable* t, size_t expected_size, upb_Arena* a) {
  int size_lg2 = upb_Log2Ceiling(_upb_entries_needed_for(expected_size));
  return init(&t->t, size_lg2, a);
}

void upb_exttable_clear(upb_exttable* t) {
  size_t bytes = upb_table_size(&t->t) * sizeof(upb_tabent);
  t->t.count = 0;
  memset((char*)t->t.entries, 0, bytes);
}

bool upb_exttable_resize(upb_exttable* t, size_t size_lg2, upb_Arena* a) {
  upb_exttable new_table;
  if (!init(&new_table.t, size_lg2, a)) return false;

  size_t i;
  for (i = begin(&t->t); i < upb_table_size(&t->t); i = next(&t->t, i)) {
    const upb_tabent* e = &t->t.entries[i];
    uint32_t hash = exthash(e->key, e->val);
    uint32_t ext_num = *(const uint32_t*)upb_value_getconstptr(e->val);
    lookupkey_t lookupkey = extkey((const void*)e->key.num, ext_num);
    insert(&new_table.t, lookupkey, e->key, e->val, hash, &exthash, &exteql);
  }

  *t = new_table;
  return true;
}

bool upb_exttable_insert(upb_exttable* t, const void* k, const uint32_t* v,
                         upb_Arena* a) {
  UPB_ASSERT(k != NULL);
  UPB_ASSERT(v != NULL);
  UPB_ASSERT(*v != 0);

  if (isfull(&t->t)) {
    if (!upb_exttable_resize(t, _upb_log2_table_size(&t->t) + 1, a)) {
      return false;
    }
  }

  lookupkey_t lookupkey = extkey(k, *v);
  upb_key key = {.num = (uintptr_t)k};
  upb_value val = upb_value_constptr(v);
  uint32_t hash = _upb_exttable_hash(k, *v);
  insert(&t->t, lookupkey, key, val, hash, &exthash, &exteql);
  return true;
}

const uint32_t* upb_exttable_lookup(const upb_exttable* t, const void* k,
                                    uint32_t ext_number) {
  uint32_t hash = _upb_exttable_hash(k, ext_number);
  upb_value val;
  if (lookup(&t->t, extkey(k, ext_number), &val, hash, &exteql)) {
    return (const uint32_t*)upb_value_getconstptr(val);
  }
  return NULL;
}

const uint32_t* upb_exttable_remove(upb_exttable* t, const void* k,
                                    uint32_t ext_number) {
  uint32_t hash = _upb_exttable_hash(k, ext_number);
  upb_value val;
  if (rm(&t->t, extkey(k, ext_number), &val, hash, &exteql)) {
    return (const uint32_t*)upb_value_getconstptr(val);
  }
  return NULL;
}

size_t upb_exttable_size(const upb_exttable* t) { return t->t.count; }

/* upb_inttable ***************************************************************/

static uint32_t inthash(upb_key key, upb_value val) {
  UPB_UNUSED(val);
  return upb_inthash(key.num);
}

static bool inteql(upb_key k1, upb_value v1, lookupkey_t k2) {
  UPB_UNUSED(v1);
  return k1.num == k2.num;
}

size_t upb_inttable_count(const upb_inttable* t) { return t->t.count; }

static void check(upb_inttable* t) {
  UPB_UNUSED(t);
#if defined(UPB_DEBUG_TABLE) && !defined(NDEBUG)
  {
    // This check is very expensive (makes inserts/deletes O(N)).
    size_t count = 0;
    intptr_t iter = UPB_INTTABLE_BEGIN;
    uintptr_t key;
    upb_value val;
    while (upb_inttable_next(t, &key, &val, &iter)) {
      UPB_ASSERT(upb_inttable_lookup(t, key, NULL));
      count++;
    }
    UPB_ASSERT(count == upb_inttable_count(t));
  }
#endif
}

static bool upb_inttable_sizedinit(upb_inttable* t, int hsize_lg2,
                                   upb_Arena* a) {
  if (!init(&t->t, hsize_lg2, a)) return false;
  check(t);
  return true;
}

bool upb_inttable_init(upb_inttable* t, upb_Arena* a) {
  return upb_inttable_sizedinit(t, 3, a);
}

bool upb_inttable_insert(upb_inttable* t, uintptr_t key, upb_value val,
                         upb_Arena* a) {
  if (isfull(&t->t)) {
    upb_table new_table;

    if (!init(&new_table, _upb_log2_table_size(&t->t) + 1, a)) {
      return false;
    }

    for (size_t i = begin(&t->t); i < upb_table_size(&t->t);
         i = next(&t->t, i)) {
      const upb_tabent* e = &t->t.entries[i];
      insert(&new_table, intkey(e->key.num), e->key, e->val,
             inthash(e->key, e->val), &inthash, &inteql);
    }

    UPB_ASSERT(t->t.count == new_table.count);

    t->t = new_table;
  }
  upb_key tabkey = {.num = key};
  insert(&t->t, intkey(key), tabkey, val, upb_inthash(key), &inthash, &inteql);
  check(t);
  return true;
}

bool upb_inttable_lookup(const upb_inttable* t, uintptr_t key, upb_value* v) {
  return lookup(&t->t, intkey(key), v, upb_inthash(key), &inteql);
}

bool upb_inttable_replace(upb_inttable* t, uintptr_t key, upb_value val) {
  upb_tabent* e =
      findentry_mutable(&t->t, intkey(key), upb_inthash(key), &inteql);
  if (!e) return false;
  e->val = val;
  return true;
}

bool upb_inttable_remove(upb_inttable* t, uintptr_t key, upb_value* val) {
  bool success = rm(&t->t, intkey(key), val, upb_inthash(key), &inteql);
  check(t);
  return success;
}

void upb_inttable_clear(upb_inttable* t) {
  size_t bytes = upb_table_size(&t->t) * sizeof(upb_tabent);
  t->t.count = 0;
  memset((char*)t->t.entries, 0, bytes);
}

bool upb_inttable_next(const upb_inttable* t, uintptr_t* key, upb_value* val,
                       intptr_t* iter) {
  upb_tabent* ent;
  if (_upb_tablenext(&t->t, &ent, iter)) {
    *key = ent->key.num;
    *val = ent->val;
    return true;
  }
  *iter = INTPTR_MAX - 1;
  return false;
}

void upb_inttable_removeiter(upb_inttable* t, intptr_t* iter) {
  removeiter(&t->t, iter);
}

void upb_inttable_setentryvalue(upb_inttable* t, intptr_t iter, upb_value v) {
  t->t.entries[iter].val = v;
}

bool upb_inttable_done(const upb_inttable* t, intptr_t iter) {
  return _upb_table_done(&t->t, iter);
}

uintptr_t upb_inttable_iter_key(const upb_inttable* t, intptr_t iter) {
  UPB_ASSERT(!upb_inttable_done(t, iter));
  return t->t.entries[iter].key.num;
}

upb_value upb_inttable_iter_value(const upb_inttable* t, intptr_t iter) {
  UPB_ASSERT(!upb_inttable_done(t, iter));
  return t->t.entries[iter].val;
}


#include <stdlib.h>

// Must be last.

static void* upb_global_allocfunc(upb_alloc* alloc, void* ptr, size_t oldsize,
                                  size_t size, size_t* actual_size) {
  UPB_UNUSED(alloc);
  UPB_UNUSED(oldsize);
  UPB_UNUSED(actual_size);
  if (size == 0) {
    free(ptr);
    return NULL;
  } else {
    return realloc(ptr, size);
  }
}

upb_alloc upb_alloc_global = {&upb_global_allocfunc};


#include <string.h>


#ifdef UPB_TRACING_ENABLED
#include <stdatomic.h>
#endif

#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Must be last.

static UPB_ATOMIC(size_t) g_max_block_size = UPB_DEFAULT_MAX_BLOCK_SIZE;

void upb_Arena_SetMaxBlockSize(size_t max) {
  UPB_ASSERT(max <= UINT32_MAX);
  upb_Atomic_Store(&g_max_block_size, max, memory_order_relaxed);
}

typedef struct upb_MemBlock {
  struct upb_MemBlock* next;
  // Size of the actual allocation.
  // Size of 0 means this is a upb_ArenaRef.
  size_t size;
  // Data follows.
} upb_MemBlock;

// A special block type that indicates a reference to another arena.
// When this arena is freed, a ref is released on the referenced arena.
// size must be 0.
typedef struct upb_ArenaRef {
  upb_MemBlock prefix;  // size is always zero
  const upb_Arena* arena;
#ifndef NDEBUG
  const struct upb_ArenaRef* next_ref;
#endif
} upb_ArenaRef;

typedef struct upb_ArenaInternal {
  // upb_alloc* together with a low bit which signals if there is an initial
  // block.
  uintptr_t block_alloc;

  // Linked list of blocks to free/cleanup.
  upb_MemBlock* blocks;

#ifndef NDEBUG
  // Stack of pointers to other arenas that this arena owns.
  // Used for debug-only ref cycle checks.
  UPB_ATOMIC(const upb_ArenaRef*) refs;
#endif

  // Size of the last block we allocated in the normal exponential scheme.
  uint32_t last_block_size;

  // A hint that grows whenever we perform a "one-off" allocation into a a
  // dedicated block. This helps us determine if these outlier blocks are
  // actually common enough that we should switch back to the normal exponential
  // scheme at the larger size.
  uint32_t size_hint;

  // All non atomic members used during allocation must be above this point, and
  // are used by _SwapIn/_SwapOut

  // Total space allocated in blocks, atomic only for SpaceAllocated
  UPB_ATOMIC(uintptr_t) space_allocated;

  // The cleanup for the allocator. This is called after all the blocks are
  // freed in an arena.
  upb_AllocCleanupFunc* upb_alloc_cleanup;

  // When multiple arenas are fused together, each arena points to a parent
  // arena (root points to itself). The root tracks how many live arenas
  // reference it.

  // The low bit is tagged:
  //   0: pointer to parent
  //   1: count, left shifted by one
  UPB_ATOMIC(uintptr_t) parent_or_count;

  // All nodes that are fused together are in a singly-linked list.
  // == NULL at end of list.
  UPB_ATOMIC(struct upb_ArenaInternal*) next;

  // - If the low bit is set, is a pointer to the tail of the list (populated
  //   for roots, set to self for roots with no fused arenas). This is best
  //   effort, and it may not always reflect the true tail, but it will always
  //   be a valid node in the list. This is useful for finding the list tail
  //   without having to walk the entire list.
  // - If the low bit is not set, is a pointer to the previous node in the list,
  //   such that a->previous_or_tail->next == a.
  UPB_ATOMIC(uintptr_t) previous_or_tail;

  // We use a different UPB_XSAN_MEMBER than the one in upb_Arena because the
  // two are distinct synchronization domains.  The upb_Arena.ptr member is
  // not published in the allocation path, so it is not synchronized with
  // respect to operations performed in this file such as Fuse, Free,
  // SpaceAllocated, etc.  This means that it is not safe to read or write
  // the upb_Arena.ptr member in those functions.
  UPB_XSAN_MEMBER
} upb_ArenaInternal;

// All public + private state for an arena.
typedef struct {
  upb_Arena head;
  upb_ArenaInternal body;
} upb_ArenaState;

typedef struct {
  upb_ArenaInternal* root;
  uintptr_t tagged_count;
} upb_ArenaRoot;

static const size_t kUpb_MemblockReserve =
    UPB_ALIGN_MALLOC(sizeof(upb_MemBlock));

static const size_t kUpb_ArenaStateReserve =
    UPB_ALIGN_MALLOC(sizeof(upb_ArenaState));

static const size_t kUpb_ArenaRefReserve =
    UPB_ALIGN_MALLOC(sizeof(upb_ArenaRef));

// Extracts the (upb_ArenaInternal*) from a (upb_Arena*)
static upb_ArenaInternal* upb_Arena_Internal(const upb_Arena* a) {
  return &((upb_ArenaState*)a)->body;
}

static bool _upb_Arena_IsTaggedRefcount(uintptr_t parent_or_count) {
  return (parent_or_count & 1) == 1;
}

static bool _upb_Arena_IsTaggedPointer(uintptr_t parent_or_count) {
  return (parent_or_count & 1) == 0;
}

static uintptr_t _upb_Arena_RefCountFromTagged(uintptr_t parent_or_count) {
  UPB_ASSERT(_upb_Arena_IsTaggedRefcount(parent_or_count));
  return parent_or_count >> 1;
}

static uintptr_t _upb_Arena_TaggedFromRefcount(uintptr_t refcount) {
  uintptr_t parent_or_count = (refcount << 1) | 1;
  UPB_ASSERT(_upb_Arena_IsTaggedRefcount(parent_or_count));
  return parent_or_count;
}

static upb_ArenaInternal* _upb_Arena_PointerFromTagged(
    uintptr_t parent_or_count) {
  UPB_ASSERT(_upb_Arena_IsTaggedPointer(parent_or_count));
  return (upb_ArenaInternal*)parent_or_count;
}

static uintptr_t _upb_Arena_TaggedFromPointer(upb_ArenaInternal* ai) {
  uintptr_t parent_or_count = (uintptr_t)ai;
  UPB_ASSERT(_upb_Arena_IsTaggedPointer(parent_or_count));
  return parent_or_count;
}

static bool _upb_Arena_IsTaggedTail(uintptr_t previous_or_tail) {
  return (previous_or_tail & 1) == 1;
}

static bool _upb_Arena_IsTaggedPrevious(uintptr_t previous_or_tail) {
  return (previous_or_tail & 1) == 0;
}

static upb_ArenaInternal* _upb_Arena_TailFromTagged(
    uintptr_t previous_or_tail) {
  UPB_ASSERT(_upb_Arena_IsTaggedTail(previous_or_tail));
  return (upb_ArenaInternal*)(previous_or_tail ^ 1);
}

static uintptr_t _upb_Arena_TaggedFromTail(upb_ArenaInternal* tail) {
  uintptr_t previous_or_tail = (uintptr_t)tail | 1;
  UPB_ASSERT(_upb_Arena_IsTaggedTail(previous_or_tail));
  return previous_or_tail;
}

static upb_ArenaInternal* _upb_Arena_PreviousFromTagged(
    uintptr_t previous_or_tail) {
  UPB_ASSERT(_upb_Arena_IsTaggedPrevious(previous_or_tail));
  return (upb_ArenaInternal*)previous_or_tail;
}

static uintptr_t _upb_Arena_TaggedFromPrevious(upb_ArenaInternal* ai) {
  uintptr_t previous = (uintptr_t)ai;
  UPB_ASSERT(_upb_Arena_IsTaggedPrevious(previous));
  return previous;
}

static upb_alloc* _upb_ArenaInternal_BlockAlloc(upb_ArenaInternal* ai) {
  return (upb_alloc*)(ai->block_alloc & ~0x1);
}

static uintptr_t _upb_Arena_MakeBlockAlloc(upb_alloc* alloc, bool has_initial) {
  uintptr_t alloc_uint = (uintptr_t)alloc;
  UPB_ASSERT((alloc_uint & 1) == 0);
  return alloc_uint | (has_initial ? 1 : 0);
}

static bool _upb_ArenaInternal_HasInitialBlock(upb_ArenaInternal* ai) {
  return ai->block_alloc & 0x1;
}

#ifdef UPB_TRACING_ENABLED
static void (*_init_arena_trace_handler)(const upb_Arena*, size_t size) = NULL;
static void (*_fuse_arena_trace_handler)(const upb_Arena*,
                                         const upb_Arena*) = NULL;
static void (*_free_arena_trace_handler)(const upb_Arena*) = NULL;

void upb_Arena_SetTraceHandler(
    void (*initArenaTraceHandler)(const upb_Arena*, size_t size),
    void (*fuseArenaTraceHandler)(const upb_Arena*, const upb_Arena*),
    void (*freeArenaTraceHandler)(const upb_Arena*)) {
  _init_arena_trace_handler = initArenaTraceHandler;
  _fuse_arena_trace_handler = fuseArenaTraceHandler;
  _free_arena_trace_handler = freeArenaTraceHandler;
}

void upb_Arena_LogInit(const upb_Arena* arena, size_t size) {
  if (_init_arena_trace_handler) {
    _init_arena_trace_handler(arena, size);
  }
}
void upb_Arena_LogFuse(const upb_Arena* arena1, const upb_Arena* arena2) {
  if (_fuse_arena_trace_handler) {
    _fuse_arena_trace_handler(arena1, arena2);
  }
}
void upb_Arena_LogFree(const upb_Arena* arena) {
  if (_free_arena_trace_handler) {
    _free_arena_trace_handler(arena);
  }
}
#endif  // UPB_TRACING_ENABLED

// If the param a is already the root, provides no memory order of refcount.
// If it has a parent, then acquire memory order is provided for both the root
// and the refcount. Thread safe.
static upb_ArenaRoot _upb_Arena_FindRoot(upb_ArenaInternal* ai) {
  uintptr_t poc = upb_Atomic_Load(&ai->parent_or_count, memory_order_relaxed);
  if (_upb_Arena_IsTaggedRefcount(poc)) {
    // Fast, relaxed path - arenas that have never been fused to a parent only
    // need relaxed memory order, since they're returning themselves and the
    // refcount.
    return (upb_ArenaRoot){.root = ai, .tagged_count = poc};
  }
  // Slow path needs acquire order; reloading is cheaper than a fence on ARM
  // (LDA vs DMB ISH). Even though this is a reread, we know it must be a tagged
  // pointer because if this Arena isn't a root, it can't ever become one.
  poc = upb_Atomic_Load(&ai->parent_or_count, memory_order_acquire);
  do {
    upb_ArenaInternal* next = _upb_Arena_PointerFromTagged(poc);
    UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(next));
    UPB_ASSERT(ai != next);
    poc = upb_Atomic_Load(&next->parent_or_count, memory_order_acquire);

    if (_upb_Arena_IsTaggedPointer(poc)) {
      // To keep complexity down, we lazily collapse levels of the tree.  This
      // keeps it flat in the final case, but doesn't cost much incrementally.
      //
      // Path splitting keeps time complexity down, see:
      //   https://en.wikipedia.org/wiki/Disjoint-set_data_structure
      UPB_ASSERT(ai != _upb_Arena_PointerFromTagged(poc));
      upb_Atomic_Store(&ai->parent_or_count, poc, memory_order_release);
    }
    ai = next;
  } while (_upb_Arena_IsTaggedPointer(poc));
  return (upb_ArenaRoot){.root = ai, .tagged_count = poc};
}

uintptr_t upb_Arena_SpaceAllocated(const upb_Arena* arena,
                                   size_t* fused_count) {
  upb_ArenaInternal* ai = upb_Arena_Internal(arena);
  uintptr_t memsize = 0;
  size_t local_fused_count = 0;
  // Our root would get updated by any racing fuses before our target arena
  // became reachable from the root via the linked list; in order to preserve
  // monotonic output (any arena counted by a previous invocation is counted by
  // this one), we instead iterate forwards and backwards so that we only see
  // the results of completed fuses.
  uintptr_t previous_or_tail =
      upb_Atomic_Load(&ai->previous_or_tail, memory_order_acquire);
  while (_upb_Arena_IsTaggedPrevious(previous_or_tail)) {
    upb_ArenaInternal* previous =
        _upb_Arena_PreviousFromTagged(previous_or_tail);
    UPB_ASSERT(previous != ai);
    UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(previous));
    // Unfortunate macro behavior; prior to C11 when using nonstandard atomics
    // this returns a void* and can't be used with += without an intermediate
    // conversion to an integer.
    // Relaxed is safe - no subsequent reads depend this one
    uintptr_t allocated =
        upb_Atomic_Load(&previous->space_allocated, memory_order_relaxed);
    memsize += allocated;
    previous_or_tail =
        upb_Atomic_Load(&previous->previous_or_tail, memory_order_acquire);
    local_fused_count++;
  }
  while (ai != NULL) {
    UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(ai));
    // Unfortunate macro behavior; prior to C11 when using nonstandard atomics
    // this returns a void* and can't be used with += without an intermediate
    // conversion to an integer.
    // Relaxed is safe - no subsequent reads depend this one
    uintptr_t allocated =
        upb_Atomic_Load(&ai->space_allocated, memory_order_relaxed);
    memsize += allocated;
    ai = upb_Atomic_Load(&ai->next, memory_order_acquire);
    local_fused_count++;
  }

  if (fused_count) *fused_count = local_fused_count;
  return memsize;
}

uint32_t upb_Arena_DebugRefCount(const upb_Arena* a) {
  uintptr_t tagged = _upb_Arena_FindRoot(upb_Arena_Internal(a)).tagged_count;
  return (uint32_t)_upb_Arena_RefCountFromTagged(tagged);
}

#if UPB_ENABLE_REF_CYCLE_CHECKS

bool upb_Arena_HasRefChain(const upb_Arena* from, const upb_Arena* to) {
  upb_ArenaInternal* ai = upb_Arena_Internal(from);
  upb_ArenaInternal* current;

  if (upb_Arena_IsFused(from, to)) return true;

  // 1. Traverse backward to the start of a consistent segment.
  uintptr_t previous_or_tail =
      upb_Atomic_Load(&ai->previous_or_tail, memory_order_acquire);
  while (_upb_Arena_IsTaggedPrevious(previous_or_tail)) {
    ai = _upb_Arena_PreviousFromTagged(previous_or_tail);
    UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(ai));
    previous_or_tail =
        upb_Atomic_Load(&ai->previous_or_tail, memory_order_acquire);
  }

  // 2. Traverse forward through all arenas in the fuse group.
  current = ai;
  while (current != NULL) {
    UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(current));
    const upb_ArenaRef* ref =
        upb_Atomic_Load(&current->refs, memory_order_acquire);
    while (ref != NULL) {
      if (ref->arena == to || upb_Arena_HasRefChain(ref->arena, to)) {
        return true;
      }
      ref = ref->next_ref;
    }
    current = upb_Atomic_Load(&current->next, memory_order_acquire);
  }
  return false;
}

#endif

static void* _upb_Arena_AllocBlockInternal(upb_alloc* alloc, size_t* size) {
  size_t sz = *size + kUpb_MemblockReserve;
  upb_SizedPtr alloc_result = upb_SizeReturningMalloc(alloc, sz);
  if (!alloc_result.p) return NULL;
  upb_MemBlock* block = alloc_result.p;
  block->size = alloc_result.n;
  *size = alloc_result.n - kUpb_MemblockReserve;
  return UPB_PTR_AT(block, kUpb_MemblockReserve, void);
}

void* UPB_PRIVATE(_upb_Arena_AllocBlock)(upb_Arena* a, size_t* size) {
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
#if UPB_HWASAN
  *size = UPB_ALIGN_UP(*size, UPB_MALLOC_ALIGN);
#endif
  void* ret =
      _upb_Arena_AllocBlockInternal(_upb_ArenaInternal_BlockAlloc(ai), size);
#if UPB_HWASAN
  if (ret) *size = UPB_ALIGN_DOWN(*size, UPB_MALLOC_ALIGN);
#endif
  return ret;
}

void UPB_PRIVATE(_upb_Arena_AddBlock)(upb_Arena* a, void* block) {
  upb_MemBlock* b = (upb_MemBlock*)((char*)block - kUpb_MemblockReserve);
  upb_ArenaInternal* ai = upb_Arena_Internal(a);

  // Atomic add not required here, as threads won't race allocating blocks, plus
  // atomic fetch-add is slower than load/add/store on arm devices compiled
  // targeting pre-v8.1. Relaxed order is safe as nothing depends on order of
  // size allocated.
  uintptr_t old_space_allocated =
      upb_Atomic_Load(&ai->space_allocated, memory_order_relaxed);
  upb_Atomic_Store(&ai->space_allocated, old_space_allocated + b->size,
                   memory_order_relaxed);

  b->next = ai->blocks;
  ai->blocks = b;
}

void UPB_PRIVATE(_upb_Arena_UseBlock)(upb_Arena* a, void* ptr, size_t size) {
  UPB_ASSERT(UPB_ALIGN_DOWN((uintptr_t)ptr, UPB_MALLOC_ALIGN) ==
             (uintptr_t)ptr);
#if UPB_HWASAN
  size = UPB_ALIGN_DOWN(size, UPB_MALLOC_ALIGN);
#endif
  if (size <= UPB_PRIVATE(_upb_ArenaHas)(a)) return;
  a->UPB_ONLYBITS(ptr) = ptr;
  a->UPB_ONLYBITS(end) = UPB_PTR_AT(ptr, size, char);
  UPB_PRIVATE(upb_Xsan_PoisonRegion)(ptr, size);
}

static bool _upb_Arena_WouldReduceFreeSpace(upb_Arena* a, size_t span,
                                            size_t block_size) {
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  size_t current_free =
      ai->blocks ? a->UPB_PRIVATE(end) - a->UPB_PRIVATE(ptr) : 0;
  size_t future_free = block_size - span;
  return current_free >= future_free;
}

size_t UPB_PRIVATE(_upb_Arena_NextBlockSize)(upb_Arena* a, size_t span,
                                             bool* one_off) {
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  bool is_one_off = false;

  size_t max_block_size =
      upb_Atomic_Load(&g_max_block_size, memory_order_relaxed);
  size_t block_size = UPB_MIN(ai->last_block_size * 2, max_block_size);

  if (span > block_size) {
    block_size = UPB_MIN(ai->size_hint * 2, max_block_size);
    if (span > block_size) {
      is_one_off = true;
    }
  }

  is_one_off =
      is_one_off || _upb_Arena_WouldReduceFreeSpace(a, span, block_size);

  if (is_one_off) {
    block_size = span;
  }

  if (one_off) *one_off = is_one_off;
  return block_size;
}

void UPB_PRIVATE(_upb_Arena_UpdateGrowthState)(upb_Arena* a, size_t span,
                                               size_t block_size,
                                               bool one_off) {
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  size_t max_block_size =
      upb_Atomic_Load(&g_max_block_size, memory_order_relaxed);
  if (one_off) {
    ai->size_hint = UPB_MIN(ai->size_hint + (span >> 1), max_block_size >> 1);
  } else {
    ai->last_block_size = UPB_MIN(block_size, UINT32_MAX);
    ai->size_hint = ai->last_block_size;
  }
}

// Fulfills the allocation request by allocating a new block. Returns NULL on
// allocation failure.
void* UPB_PRIVATE(_upb_Arena_SlowMalloc)(upb_Arena* a, size_t span) {
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  if (!_upb_ArenaInternal_BlockAlloc(ai)) return NULL;

  bool one_off = false;
  size_t block_size = UPB_PRIVATE(_upb_Arena_NextBlockSize)(a, span, &one_off);

  void* block = UPB_PRIVATE(_upb_Arena_AllocBlock)(a, &block_size);
  if (!block) return NULL;
  UPB_PRIVATE(_upb_Arena_AddBlock)(a, block);
  size_t size = span - UPB_PRIVATE(kUpb_Asan_GuardSize);

  // Recheck size, in case the allocator gave us a much larger block than we
  // requested and we want to make it the new allocating region.
  if (one_off && !_upb_Arena_WouldReduceFreeSpace(a, span, block_size)) {
    one_off = false;
  }

  UPB_PRIVATE(_upb_Arena_UpdateGrowthState)(a, span, block_size, one_off);

  if (one_off) {
    UPB_PRIVATE(upb_Xsan_PoisonRegion)(UPB_PTR_AT(block, size, void),
                                       block_size - size);
    return block;
  } else {
    UPB_PRIVATE(_upb_Arena_UseBlock)(a, block, block_size);
    UPB_ASSERT(UPB_PRIVATE(_upb_ArenaHas)(a) >= span);
    return upb_Arena_Malloc(a, size);
  }
}

static upb_Arena* _upb_Arena_InitSlow(upb_alloc* alloc, size_t first_size) {
  if (!alloc) return NULL;

  // We need to malloc the initial block.
  size_t block_size = kUpb_ArenaStateReserve +
                      UPB_MAX(256, UPB_ALIGN_MALLOC(first_size) +
                                       UPB_PRIVATE(kUpb_Asan_GuardSize));
  void* block = _upb_Arena_AllocBlockInternal(alloc, &block_size);
  if (!block) return NULL;

  // Initialize the arena state in the first block. We "borrow" the memory from
  // the block, because we can't yet call upb_Arena_Malloc.
  upb_ArenaState* a = block;
  a->body.block_alloc = _upb_Arena_MakeBlockAlloc(alloc, 0);
  a->body.last_block_size = UPB_MIN(block_size, UINT32_MAX);
  a->body.size_hint = UPB_MIN(block_size, UINT32_MAX);
  a->head.UPB_ONLYBITS(ptr) = NULL;
  a->head.UPB_ONLYBITS(end) = NULL;

  upb_Atomic_Init(&a->body.parent_or_count, _upb_Arena_TaggedFromRefcount(1));
  upb_Atomic_Init(&a->body.next, NULL);
  upb_Atomic_Init(&a->body.previous_or_tail,
                  _upb_Arena_TaggedFromTail(&a->body));
  upb_Atomic_Init(&a->body.space_allocated, 0);
  a->body.blocks = NULL;
#ifndef NDEBUG
  a->body.refs = NULL;
#endif
  a->body.upb_alloc_cleanup = NULL;
  UPB_PRIVATE(upb_Xsan_Init)(UPB_XSAN(&a->body));

  UPB_PRIVATE(_upb_Arena_AddBlock)(&a->head, block);

  block = UPB_PTR_AT(block, kUpb_ArenaStateReserve, void);
  block_size -= kUpb_ArenaStateReserve;
  UPB_PRIVATE(_upb_Arena_UseBlock)(&a->head, block, block_size);

  return &a->head;
}

upb_Arena* upb_Arena_Init(void* mem, size_t n, upb_alloc* alloc) {
  UPB_STATIC_ASSERT(UPB_ARENA_SIZE_HACK >= sizeof(upb_ArenaState),
                    "Need to update UPB_ARENA_SIZE_HACK");
  upb_ArenaState* a;

  if (mem) {
    /* Align initial pointer up so that we return properly-aligned pointers. */
    void* aligned = (void*)UPB_ALIGN_MALLOC((uintptr_t)mem);
    size_t delta = (uintptr_t)aligned - (uintptr_t)mem;
    n = delta <= n ? n - delta : 0;
    mem = aligned;
  }
  if (UPB_UNLIKELY(n < sizeof(upb_ArenaState) || !mem)) {
    upb_Arena* ret = _upb_Arena_InitSlow(alloc, mem ? 0 : n);
#ifdef UPB_TRACING_ENABLED
    upb_Arena_LogInit(ret, n);
#endif
    return ret;
  }

  a = mem;

  upb_Atomic_Init(&a->body.parent_or_count, _upb_Arena_TaggedFromRefcount(1));
  upb_Atomic_Init(&a->body.next, NULL);
  upb_Atomic_Init(&a->body.previous_or_tail,
                  _upb_Arena_TaggedFromTail(&a->body));
  upb_Atomic_Init(&a->body.space_allocated, 0);
  a->body.blocks = NULL;
#ifndef NDEBUG
  a->body.refs = NULL;
#endif
  a->body.size_hint = 128;
  a->body.last_block_size = 128;
  a->body.upb_alloc_cleanup = NULL;
  a->body.block_alloc = _upb_Arena_MakeBlockAlloc(alloc, 1);
  a->head.UPB_PRIVATE(ptr) = (void*)UPB_ALIGN_MALLOC((uintptr_t)(a + 1));
  a->head.UPB_PRIVATE(end) = UPB_PTR_AT(mem, n, char);
  UPB_PRIVATE(upb_Xsan_Init)(UPB_XSAN(&a->body));
#ifdef UPB_TRACING_ENABLED
  upb_Arena_LogInit(&a->head, n);
#endif
  return &a->head;
}

static void _upb_Arena_DoFree(upb_ArenaInternal* ai) {
  UPB_ASSERT(_upb_Arena_RefCountFromTagged(ai->parent_or_count) == 1);
  while (ai != NULL) {
    UPB_PRIVATE(upb_Xsan_AccessReadWrite)(UPB_XSAN(ai));
    // Load first since arena itself is likely from one of its blocks. Relaxed
    // order is safe because fused arena ordering is provided by the reference
    // count, and fuse is not permitted to race with the final decrement.
    upb_ArenaInternal* next_arena =
        (upb_ArenaInternal*)upb_Atomic_Load(&ai->next, memory_order_relaxed);
    // Freeing may have memory barriers that confuse tsan, so assert immediately
    // after load here
    if (next_arena) {
      UPB_PRIVATE(upb_Xsan_AccessReadWrite)(UPB_XSAN(next_arena));
    }
    upb_alloc* block_alloc = _upb_ArenaInternal_BlockAlloc(ai);
    upb_MemBlock* block = ai->blocks;
    upb_AllocCleanupFunc* alloc_cleanup = ai->upb_alloc_cleanup;
    while (block != NULL) {
      // Load first since we are deleting block.
      upb_MemBlock* next_block = block->next;
      if (block->size == 0) {
        // If the block is an arena ref, then we need to release our ref on the
        // referenced arena.
        upb_ArenaRef* ref = (upb_ArenaRef*)block;
        upb_Arena_DecRefFor((upb_Arena*)ref->arena, ai);
      } else {
        upb_free_sized(block_alloc, block, block->size);
      }
      block = next_block;
    }
    if (alloc_cleanup != NULL) {
      alloc_cleanup(block_alloc);
    }
    ai = next_arena;
  }
}

void upb_Arena_Free(upb_Arena* a) {
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  // Cannot be replaced with _upb_Arena_FindRoot, as that provides only a
  // relaxed read of the refcount if ai is already the root.
  uintptr_t poc = upb_Atomic_Load(&ai->parent_or_count, memory_order_acquire);
retry:
  while (_upb_Arena_IsTaggedPointer(poc)) {
    ai = _upb_Arena_PointerFromTagged(poc);
    UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(ai));
    poc = upb_Atomic_Load(&ai->parent_or_count, memory_order_acquire);
  }

  // compare_exchange or fetch_sub are RMW operations, which are more
  // expensive then direct loads.  As an optimization, we only do RMW ops
  // when we need to update things for other threads to see.
  if (poc == _upb_Arena_TaggedFromRefcount(1)) {
#ifdef UPB_TRACING_ENABLED
    upb_Arena_LogFree(a);
#endif
    _upb_Arena_DoFree(ai);
    return;
  }

  if (upb_Atomic_CompareExchangeWeak(
          &ai->parent_or_count, &poc,
          _upb_Arena_TaggedFromRefcount(_upb_Arena_RefCountFromTagged(poc) - 1),
          memory_order_release, memory_order_acquire)) {
    // We were >1 and we decremented it successfully, so we are done.
    return;
  }

  // We failed our update, so someone has done something, retry the whole
  // process, but the failed exchange reloaded `poc` for us.
  goto retry;
}

// Logically performs the following operation, in a way that is safe against
// racing fuses:
//   ret = TAIL(parent)
//   ret->next = child
//   return ret
//
// The caller is therefore guaranteed that ret->next == child.
static upb_ArenaInternal* _upb_Arena_LinkForward(
    upb_ArenaInternal* const parent, upb_ArenaInternal* child) {
  UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(parent));
  uintptr_t parent_previous_or_tail =
      upb_Atomic_Load(&parent->previous_or_tail, memory_order_acquire);

  // Optimization: use parent->previous_or_tail to skip to TAIL(parent) in O(1)
  // time when possible. This is the common case because we just fused into
  // parent, suggesting that it should be a root with a cached tail.
  //
  // However, if there was a racing fuse, parent may no longer be a root, in
  // which case we need to walk the entire list to find the tail. The tail
  // pointer is also not guaranteed to be the true tail, so even when the
  // optimization is taken, we still need to walk list nodes to find the true
  // tail.
  upb_ArenaInternal* parent_tail =
      _upb_Arena_IsTaggedTail(parent_previous_or_tail)
          ? _upb_Arena_TailFromTagged(parent_previous_or_tail)
          : parent;

  UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(parent_tail));
  upb_ArenaInternal* parent_tail_next =
      upb_Atomic_Load(&parent_tail->next, memory_order_acquire);

  do {
    // Walk the list to find the true tail (a node with next == NULL).
    while (parent_tail_next != NULL) {
      parent_tail = parent_tail_next;
      UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(parent_tail));
      parent_tail_next =
          upb_Atomic_Load(&parent_tail->next, memory_order_acquire);
    }
  } while (!upb_Atomic_CompareExchangeWeak(  // Replace a NULL next with child.
      &parent_tail->next, &parent_tail_next, child, memory_order_release,
      memory_order_acquire));

  return parent_tail;
}

// Updates parent->previous_or_tail = child->previous_or_tail in hopes that the
// latter represents the true tail of the newly-combined list.
//
// This is a best-effort operation that may set the tail to a stale value, and
// may fail to update the tail at all.
void _upb_Arena_UpdateParentTail(upb_ArenaInternal* parent,
                                 upb_ArenaInternal* child) {
  // We are guaranteed that child->previous_or_tail is tagged, because we have
  // just transitioned child from root -> non-root, which is an exclusive
  // operation that can only happen once. So we are the exclusive updater of
  // child->previous_or_tail that can transition it from tagged to untagged.
  //
  // However, we are not guaranteed that child->previous_or_tail is the true
  // tail.  A racing fuse may have appended to child's list but not yet updated
  // child->previous_or_tail.
  uintptr_t child_previous_or_tail =
      upb_Atomic_Load(&child->previous_or_tail, memory_order_acquire);
  upb_ArenaInternal* new_parent_tail =
      _upb_Arena_TailFromTagged(child_previous_or_tail);
  UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(new_parent_tail));

  // If another thread fused with parent, such that it is no longer a root,
  // don't overwrite their previous pointer with our tail. Relaxed order is fine
  // here as we only inspect the tag bit.
  uintptr_t parent_previous_or_tail =
      upb_Atomic_Load(&parent->previous_or_tail, memory_order_relaxed);
  if (_upb_Arena_IsTaggedTail(parent_previous_or_tail)) {
    upb_Atomic_CompareExchangeStrong(
        &parent->previous_or_tail, &parent_previous_or_tail,
        _upb_Arena_TaggedFromTail(new_parent_tail), memory_order_release,
        memory_order_relaxed);
  }
}

static void _upb_Arena_LinkBackward(upb_ArenaInternal* child,
                                    upb_ArenaInternal* old_parent_tail) {
  // Link child to parent going backwards, for SpaceAllocated.  This transitions
  // child->previous_or_tail from tail (tagged) to previous (untagged), after
  // which its value is immutable.
  //
  // - We are guaranteed that no other threads are also attempting to perform
  //   this transition (tail -> previous), because we just updated
  //   old_parent_tail->next from NULL to non-NULL, an exclusive operation that
  //   can only happen once.
  //
  // - _upb_Arena_UpdateParentTail() uses CAS to ensure that it
  //    does not perform the reverse transition (previous -> tail).
  //
  // - We are guaranteed that old_parent_tail is the correct "previous" pointer,
  //   even in the presence of racing fuses that are adding more nodes to the
  //   list, because _upb_Arena_LinkForward() guarantees that:
  //       old_parent_tail->next == child.
  upb_Atomic_Store(&child->previous_or_tail,
                   _upb_Arena_TaggedFromPrevious(old_parent_tail),
                   memory_order_release);
}

static void _upb_Arena_DoFuseArenaLists(upb_ArenaInternal* const parent,
                                        upb_ArenaInternal* child) {
  upb_ArenaInternal* old_parent_tail = _upb_Arena_LinkForward(parent, child);
  _upb_Arena_UpdateParentTail(parent, child);
  _upb_Arena_LinkBackward(child, old_parent_tail);
}

void upb_Arena_SetAllocCleanup(upb_Arena* a, upb_AllocCleanupFunc* func) {
  UPB_PRIVATE(upb_Xsan_AccessReadWrite)(UPB_XSAN(a));
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  UPB_ASSERT(ai->upb_alloc_cleanup == NULL);
  ai->upb_alloc_cleanup = func;
}

// Thread safe.
static upb_ArenaInternal* _upb_Arena_DoFuse(upb_ArenaInternal** ai1,
                                            upb_ArenaInternal** ai2,
                                            uintptr_t* ref_delta) {
  // `parent_or_count` has two distinct modes
  // -  parent pointer mode
  // -  refcount mode
  //
  // In parent pointer mode, it may change what pointer it refers to in the
  // tree, but it will always approach a root.  Any operation that walks the
  // tree to the root may collapse levels of the tree concurrently.
  upb_ArenaRoot r1 = _upb_Arena_FindRoot(*ai1);
  upb_ArenaRoot r2 = _upb_Arena_FindRoot(*ai2);

  if (r1.root == r2.root) return r1.root;  // Already fused.

  *ai1 = r1.root;
  *ai2 = r2.root;

  // Avoid cycles by always fusing into the root with the lower address.
  if ((uintptr_t)r1.root > (uintptr_t)r2.root) {
    upb_ArenaRoot tmp = r1;
    r1 = r2;
    r2 = tmp;
  }

  // The moment we install `r1` as the parent for `r2` all racing frees may
  // immediately begin decrementing `r1`'s refcount (including pending
  // increments to that refcount and their frees!).  We need to add `r2`'s refs
  // now, so that `r1` can withstand any unrefs that come from r2.
  //
  // Note that while it is possible for `r2`'s refcount to increase
  // asynchronously, we will not actually do the reparenting operation below
  // unless `r2`'s refcount is unchanged from when we read it.
  //
  // Note that we may have done this previously, either to this node or a
  // different node, during a previous and failed DoFuse() attempt. But we will
  // not lose track of these refs because we always add them to our overall
  // delta.
  uintptr_t r2_untagged_count = r2.tagged_count & ~1;
  uintptr_t with_r2_refs = r1.tagged_count + r2_untagged_count;
  if (!upb_Atomic_CompareExchangeStrong(
          &r1.root->parent_or_count, &r1.tagged_count, with_r2_refs,
          memory_order_release, memory_order_acquire)) {
    return NULL;
  }

  // Perform the actual fuse by removing the refs from `r2` and swapping in the
  // parent pointer.
  if (!upb_Atomic_CompareExchangeStrong(
          &r2.root->parent_or_count, &r2.tagged_count,
          _upb_Arena_TaggedFromPointer(r1.root), memory_order_release,
          memory_order_acquire)) {
    // We'll need to remove the excess refs we added to r1 previously.
    *ref_delta += r2_untagged_count;
    return NULL;
  }

  // Now that the fuse has been performed (and can no longer fail) we need to
  // append `r2` to `r1`'s linked list.
  _upb_Arena_DoFuseArenaLists(r1.root, r2.root);
  return r1.root;
}

// Thread safe.
static bool _upb_Arena_FixupRefs(upb_ArenaInternal* new_root,
                                 uintptr_t ref_delta) {
  if (ref_delta == 0) return true;  // No fixup required.
  // Relaxed order is safe here as if the value is a pointer, we don't deref it
  // or publish it anywhere else. The refcount does provide memory order
  // between allocations on arenas and the eventual free and thus normally
  // requires acquire/release; but in this case any edges provided by the refs
  // we are cleaning up were already provided by the fuse operation itself. It's
  // not valid for a decrement that could cause the overall fused arena to reach
  // a zero refcount to race with this function, as that could result in a
  // use-after-free anyway.
  uintptr_t poc =
      upb_Atomic_Load(&new_root->parent_or_count, memory_order_relaxed);
  if (_upb_Arena_IsTaggedPointer(poc)) return false;
  uintptr_t with_refs = poc - ref_delta;
  UPB_ASSERT(!_upb_Arena_IsTaggedPointer(with_refs));
  // Relaxed order on success is safe here, for the same reasons as the relaxed
  // read above. Relaxed order is safe on failure because the updated value is
  // stored in a local variable which goes immediately out of scope; the retry
  // loop will reread what it needs with proper memory order.
  return upb_Atomic_CompareExchangeStrong(&new_root->parent_or_count, &poc,
                                          with_refs, memory_order_relaxed,
                                          memory_order_relaxed);
}

bool upb_Arena_Fuse(const upb_Arena* a1, const upb_Arena* a2) {
  if (a1 == a2) return true;  // trivial fuse

#ifdef UPB_TRACING_ENABLED
  upb_Arena_LogFuse(a1, a2);
#endif

  upb_ArenaInternal* ai1 = upb_Arena_Internal(a1);
  upb_ArenaInternal* ai2 = upb_Arena_Internal(a2);

  // Do not fuse initial blocks since we cannot lifetime extend them.
  // Any other fuse scenario is allowed.
  if (_upb_ArenaInternal_HasInitialBlock(ai1) ||
      _upb_ArenaInternal_HasInitialBlock(ai2)) {
    return false;
  }

  // The number of refs we ultimately need to transfer to the new root.
  uintptr_t ref_delta = 0;
  while (true) {
    upb_ArenaInternal* new_root = _upb_Arena_DoFuse(&ai1, &ai2, &ref_delta);
    if (new_root != NULL && _upb_Arena_FixupRefs(new_root, ref_delta)) {
#if UPB_ENABLE_REF_CYCLE_CHECKS
      UPB_ASSERT(!upb_Arena_HasRefChain(a1, a2));
#endif
      return true;
    }
  }
}

bool upb_Arena_IsFused(const upb_Arena* a, const upb_Arena* b) {
  if (a == b) return true;  // trivial fuse
  upb_ArenaInternal* ra = _upb_Arena_FindRoot(upb_Arena_Internal(a)).root;
  upb_ArenaInternal* rb = upb_Arena_Internal(b);
  while (true) {
    rb = _upb_Arena_FindRoot(rb).root;
    if (ra == rb) return true;
    upb_ArenaInternal* tmp = _upb_Arena_FindRoot(ra).root;
    if (ra == tmp) return false;
    // a's root changed since we last checked.  Retry.
    ra = tmp;
  }
}

bool upb_Arena_IncRefFor(const upb_Arena* a, const void* owner) {
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  if (_upb_ArenaInternal_HasInitialBlock(ai)) return false;
  upb_ArenaRoot r;
  r.root = ai;

retry:
  r = _upb_Arena_FindRoot(r.root);
  if (upb_Atomic_CompareExchangeWeak(
          &r.root->parent_or_count, &r.tagged_count,
          _upb_Arena_TaggedFromRefcount(
              _upb_Arena_RefCountFromTagged(r.tagged_count) + 1),
          // Relaxed order is safe on success, incrementing the refcount
          // need not perform any synchronization with the eventual free of the
          // arena - that's provided by decrements.
          memory_order_relaxed,
          // Relaxed order is safe on failure as r.tagged_count is immediately
          // overwritten by retrying the find root operation.
          memory_order_relaxed)) {
    // We incremented it successfully, so we are done.
    return true;
  }
  // We failed update due to parent switching on the arena.
  goto retry;
}

void upb_Arena_DecRefFor(const upb_Arena* a, const void* owner) {
  upb_Arena_Free((upb_Arena*)a);
}

bool upb_Arena_RefArena(upb_Arena* from, const upb_Arena* to) {
  UPB_ASSERT(!upb_Arena_IsFused(from, to));
  if (_upb_ArenaInternal_HasInitialBlock(upb_Arena_Internal(to))) {
    // We can't increment a ref to `to`, so return early.
    return false;
  }

  upb_ArenaInternal* ai = upb_Arena_Internal(from);
  upb_ArenaRef* ref = upb_Arena_Malloc(from, kUpb_ArenaRefReserve);

  if (!ref) {
    return false;
  }

  // When 'from' is freed, a ref on 'to' will be released.
  // Intentionally ignore return value, since we already check up above if this
  // call will succeed.
  bool result = upb_Arena_IncRefFor(to, from);
  UPB_ASSERT(result);

  // When we add a reference from `from` to `to`, we need to keep track of the
  // ref in the `from` arena's linked list of refs. This allows us to
  // walk all refs for `from` when `from` is freed, and thus allows us to
  // decrement the refcount on `to` when `from` is freed.
  ref->prefix.next = ai->blocks;
  ref->prefix.size = 0;
  ref->arena = to;
  ai->blocks = (upb_MemBlock*)ref;

#ifndef NDEBUG
  // Add to the dedicated list of refs.
  // This function is not thread-safe from `from`, so a simple load/store is
  // sufficient.
  ref->next_ref = upb_Atomic_Load(&ai->refs, memory_order_relaxed);
  upb_Atomic_Store(&ai->refs, ref, memory_order_release);
#endif

#if UPB_ENABLE_REF_CYCLE_CHECKS
  UPB_ASSERT(!upb_Arena_HasRefChain(to, from));  // Forbid cycles.
#endif

  return true;
}

#ifndef NDEBUG
bool upb_Arena_HasRef(const upb_Arena* from, const upb_Arena* to) {
  const upb_ArenaInternal* ai = upb_Arena_Internal(from);
  const upb_ArenaRef* ref = upb_Atomic_Load(&ai->refs, memory_order_acquire);
  while (ref != NULL) {
    if (upb_Arena_IsFused(ref->arena, to)) {
      return true;
    }
    ref = ref->next_ref;
  }
  return false;
}
#endif

upb_alloc* upb_Arena_GetUpbAlloc(upb_Arena* a) {
  UPB_PRIVATE(upb_Xsan_AccessReadOnly)(UPB_XSAN(a));
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  return _upb_ArenaInternal_BlockAlloc(ai);
}

void UPB_PRIVATE(_upb_Arena_SwapIn)(upb_Arena* des, const upb_Arena* src) {
  memcpy(des, src, offsetof(upb_ArenaState, body.space_allocated));
  upb_ArenaInternal* desi = upb_Arena_Internal(des);
  upb_ArenaInternal* srci = upb_Arena_Internal(src);
  uintptr_t new_space_allocated =
      upb_Atomic_Load(&srci->space_allocated, memory_order_relaxed);
  upb_Atomic_Store(&desi->space_allocated, new_space_allocated,
                   memory_order_relaxed);
}

void UPB_PRIVATE(_upb_Arena_SwapOut)(upb_Arena* des, const upb_Arena* src) {
  UPB_PRIVATE(_upb_Arena_SwapIn)(des, src);
}

bool _upb_Arena_WasLastAllocFromPreviousBlock(struct upb_Arena* a, void* ptr,
                                              size_t oldsize) {
  upb_ArenaInternal* ai = upb_Arena_Internal(a);
  upb_MemBlock* block = ai->blocks;
  // Skip any arena refs.
  while (block != NULL && block->size == 0) {
    block = block->next;
  }
  if (block == NULL) return false;
  char* start = UPB_PTR_AT(block, kUpb_MemblockReserve, char);

  // We don't actually track the final pointer value, so we can only check that
  // the span is within the previous block.
  return UPB_PRIVATE(upb_Xsan_PtrEq)(ptr, start) &&
         UPB_PRIVATE(_upb_Arena_AllocSpan)(oldsize) ==
             block->size - kUpb_MemblockReserve;
}

void* UPB_PRIVATE(_upb_Arena_Steal)(struct upb_Arena* a, size_t* size) {
  UPB_PRIVATE(upb_Xsan_AccessReadWrite)(UPB_XSAN(a));
  size_t have = UPB_PRIVATE(_upb_ArenaHas)(a);
  if (have < *size) return NULL;

  *size = have;
  void* ptr = a->UPB_ONLYBITS(ptr);
  a->UPB_ONLYBITS(ptr) += have;

  return UPB_PRIVATE(upb_Xsan_NewUnpoisonedRegion)(UPB_XSAN(a), ptr, have);
}

// Frees a block previously allocated with _upb_Arena_AllocBlock.
void UPB_PRIVATE(_upb_Arena_FreeBlock)(upb_Arena* a, void* block) {
  UPB_PRIVATE(upb_Xsan_AccessReadWrite)(UPB_XSAN(a));
  upb_MemBlock* b =
      UPB_PTR_AT(block, -(ptrdiff_t)kUpb_MemblockReserve, upb_MemBlock);
  upb_alloc* alloc = upb_Arena_GetUpbAlloc(a);
  upb_free_sized(alloc, b, b->size);
}


#include <stddef.h>
#include <stdint.h>


// Must be last.


#ifdef __cplusplus
extern "C" {
#endif

bool upb_Message_IsEmpty(const upb_Message* msg, const upb_MiniTable* m) {
  if (upb_Message_ExtensionCount(msg)) return false;

  const upb_MiniTableField* f;
  upb_MessageValue v;
  size_t iter = kUpb_BaseField_Begin;
  return !UPB_PRIVATE(_upb_Message_NextBaseField)(msg, m, &f, &v, &iter);
}

static bool _upb_Array_IsEqual(const upb_Array* arr1, const upb_Array* arr2,
                               upb_CType ctype, const upb_MiniTable* m,
                               int options) {
  // Check for trivial equality.
  if (arr1 == arr2) return true;

  // Must have identical element counts.
  const size_t size1 = arr1 ? upb_Array_Size(arr1) : 0;
  const size_t size2 = arr2 ? upb_Array_Size(arr2) : 0;
  if (size1 != size2) return false;

  for (size_t i = 0; i < size1; i++) {
    const upb_MessageValue val1 = upb_Array_Get(arr1, i);
    const upb_MessageValue val2 = upb_Array_Get(arr2, i);

    if (!upb_MessageValue_IsEqual(val1, val2, ctype, m, options)) return false;
  }

  return true;
}

static bool _upb_Map_IsEqual(const upb_Map* map1, const upb_Map* map2,
                             const upb_MiniTable* m, int options) {
  // Check for trivial equality.
  if (map1 == map2) return true;

  // Must have identical element counts.
  size_t size1 = map1 ? upb_Map_Size(map1) : 0;
  size_t size2 = map2 ? upb_Map_Size(map2) : 0;
  if (size1 != size2) return false;

  const upb_MiniTableField* f = upb_MiniTable_MapValue(m);
  const upb_MiniTable* m2_value = upb_MiniTable_SubMessage(f);
  const upb_CType ctype = upb_MiniTableField_CType(f);

  upb_MessageValue key, val1, val2;
  size_t iter = kUpb_Map_Begin;
  while (upb_Map_Next(map1, &key, &val1, &iter)) {
    if (!upb_Map_Get(map2, key, &val2)) return false;
    if (!upb_MessageValue_IsEqual(val1, val2, ctype, m2_value, options))
      return false;
  }

  return true;
}

static bool _upb_Message_BaseFieldsAreEqual(const upb_Message* msg1,
                                            const upb_Message* msg2,
                                            const upb_MiniTable* m,
                                            int options) {
  // Iterate over all base fields for each message.
  // The order will always match if the messages are equal.
  size_t iter1 = kUpb_BaseField_Begin;
  size_t iter2 = kUpb_BaseField_Begin;

  for (;;) {
    const upb_MiniTableField *f1, *f2;
    upb_MessageValue val1, val2;

    const bool got1 =
        UPB_PRIVATE(_upb_Message_NextBaseField)(msg1, m, &f1, &val1, &iter1);
    const bool got2 =
        UPB_PRIVATE(_upb_Message_NextBaseField)(msg2, m, &f2, &val2, &iter2);

    if (got1 != got2) return false;  // Must have identical field counts.
    if (!got1) return true;          // Loop termination condition.
    if (f1 != f2) return false;      // Must have identical fields set.

    const upb_MiniTable* subm = upb_MiniTable_SubMessage(f1);
    const upb_CType ctype = upb_MiniTableField_CType(f1);

    bool eq;
    switch (UPB_PRIVATE(_upb_MiniTableField_Mode)(f1)) {
      case kUpb_FieldMode_Array:
        eq = _upb_Array_IsEqual(val1.array_val, val2.array_val, ctype, subm,
                                options);
        break;
      case kUpb_FieldMode_Map:
        eq = _upb_Map_IsEqual(val1.map_val, val2.map_val, subm, options);
        break;
      case kUpb_FieldMode_Scalar:
        eq = upb_MessageValue_IsEqual(val1, val2, ctype, subm, options);
        break;
    }
    if (!eq) return false;
  }
}

static bool _upb_Message_ExtensionsAreEqual(const upb_Message* msg1,
                                            const upb_Message* msg2,
                                            const upb_MiniTable* m,
                                            int options) {
  const upb_MiniTableExtension* e;
  upb_MessageValue val1;

  // Iterate over all extensions for msg1, and search msg2 for each extension.
  size_t count1 = 0;
  size_t iter1 = kUpb_Message_ExtensionBegin;
  while (upb_Message_NextExtension(msg1, &e, &val1, &iter1)) {
    const upb_Extension* ext2 = UPB_PRIVATE(_upb_Message_Getext)(msg2, e);
    if (!ext2) return false;

    count1++;

    const upb_MessageValue val2 = ext2->data;
    const upb_MiniTableField* f = &e->UPB_PRIVATE(field);
    const upb_MiniTable* subm = upb_MiniTableField_IsSubMessage(f)
                                    ? upb_MiniTableExtension_GetSubMessage(e)
                                    : NULL;
    const upb_CType ctype = upb_MiniTableField_CType(f);

    bool eq;
    switch (UPB_PRIVATE(_upb_MiniTableField_Mode)(f)) {
      case kUpb_FieldMode_Array:
        eq = _upb_Array_IsEqual(val1.array_val, val2.array_val, ctype, subm,
                                options);
        break;
      case kUpb_FieldMode_Map:
        UPB_UNREACHABLE();  // Maps cannot be extensions.
        break;
      case kUpb_FieldMode_Scalar: {
        eq = upb_MessageValue_IsEqual(val1, val2, ctype, subm, options);
        break;
      }
    }
    if (!eq) return false;
  }

  // Must have identical extension counts (this catches the case where msg2
  // has extensions that msg1 doesn't).
  if (count1 != upb_Message_ExtensionCount(msg2)) return false;

  return true;
}

bool upb_Message_IsEqual(const upb_Message* msg1, const upb_Message* msg2,
                         const upb_MiniTable* m, int options) {
  if (UPB_UNLIKELY(msg1 == msg2)) return true;

  if (!_upb_Message_BaseFieldsAreEqual(msg1, msg2, m, options)) return false;
  if (!_upb_Message_ExtensionsAreEqual(msg1, msg2, m, options)) return false;

  if (!(options & kUpb_CompareOption_IncludeUnknownFields)) return true;

  // The wire encoder enforces a maximum depth of 100 so we match that here.
  return UPB_PRIVATE(_upb_Message_UnknownFieldsAreEqual)(msg1, msg2, 100) ==
         kUpb_UnknownCompareResult_Equal;
}


#include <stddef.h>
#include <stdint.h>


// Must be last.

bool upb_Message_NextExtensionReverse(const upb_Message* msg,
                                      const upb_MiniTableExtension** result,
                                      uintptr_t* iter) {
  upb_MessageValue val;
  return UPB_PRIVATE(_upb_Message_NextExtensionReverse)(msg, result, &val,
                                                        iter);
}

const upb_MiniTableExtension* upb_Message_FindExtensionByNumber(
    const upb_Message* msg, uint32_t field_number) {
  uintptr_t iter = kUpb_Message_ExtensionBegin;
  const upb_MiniTableExtension* result;
  while (upb_Message_NextExtensionReverse(msg, &result, &iter)) {
    if (upb_MiniTableExtension_Number(result) == field_number) return result;
  }
  return NULL;
}


#include <stdint.h>
#include <string.h>


// Must be last.

upb_Array* upb_Array_New(upb_Arena* a, upb_CType type) {
  const int lg2 = UPB_PRIVATE(_upb_CType_SizeLg2)(type);
  return UPB_PRIVATE(_upb_Array_New)(a, _UPB_ARRAY_DEFAULT_INITIAL_SIZE, lg2);
}

upb_MessageValue upb_Array_Get(const upb_Array* arr, size_t i) {
  UPB_ASSERT(i < upb_Array_Size(arr));
  upb_MessageValue ret;
  const char* data = upb_Array_DataPtr(arr);
  const int lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(arr);
  memcpy(&ret, data + (i << lg2), 1 << lg2);
  return ret;
}

upb_Message* upb_Array_GetMutable(upb_Array* arr, size_t i) {
  UPB_ASSERT(i < upb_Array_Size(arr));
  size_t elem_size = sizeof(upb_Message*);
  UPB_ASSERT(elem_size == (1 << UPB_PRIVATE(_upb_Array_ElemSizeLg2)(arr)));
  char* data = upb_Array_MutableDataPtr(arr);
  upb_Message* ret;
  memcpy(&ret, data + (i * elem_size), elem_size);
  UPB_ASSERT(!upb_Message_IsFrozen(ret));
  return ret;
}

void upb_Array_Set(upb_Array* arr, size_t i, upb_MessageValue val) {
  UPB_ASSERT(!upb_Array_IsFrozen(arr));
  UPB_ASSERT(i < upb_Array_Size(arr));
  char* data = upb_Array_MutableDataPtr(arr);
  const int lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(arr);
  memcpy(data + (i << lg2), &val, 1 << lg2);
}

bool upb_Array_Append(upb_Array* arr, upb_MessageValue val, upb_Arena* arena) {
  UPB_ASSERT(!upb_Array_IsFrozen(arr));
  UPB_ASSERT(arena);
  if (!UPB_PRIVATE(_upb_Array_ResizeUninitialized)(
          arr, arr->UPB_PRIVATE(size) + 1, arena)) {
    return false;
  }
  upb_Array_Set(arr, arr->UPB_PRIVATE(size) - 1, val);
  return true;
}

bool upb_Array_Copy(upb_Array* dst, const upb_Array* src, upb_Arena* arena) {
  UPB_ASSERT(dst);
  UPB_ASSERT(src);
  UPB_ASSERT(!upb_Array_IsFrozen(dst));
  if (dst == src) return true;
  size_t len = upb_Array_Size(src);
  if (!UPB_PRIVATE(_upb_Array_ResizeUninitialized)(dst, len, arena)) {
    return false;
  }
  if (len == 0) return true;
  const int lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(dst);
  const int src_lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(src);
  UPB_ASSERT(lg2 == src_lg2);
  char* dst_data = upb_Array_MutableDataPtr(dst);
  const char* src_data = upb_Array_DataPtr(src);
  memcpy(dst_data, src_data, len << lg2);
  return true;
}

bool upb_Array_AppendAll(upb_Array* dst, const upb_Array* src,
                         upb_Arena* arena) {
  UPB_ASSERT(!upb_Array_IsFrozen(dst));
  UPB_ASSERT(src);
  size_t src_len = upb_Array_Size(src);
  if (src_len == 0) return true;
  size_t dst_len = upb_Array_Size(dst);
  size_t len = dst_len + src_len;
  if (UPB_UNLIKELY(len < dst_len)) return false;
  if (!UPB_PRIVATE(_upb_Array_ResizeUninitialized)(dst, len, arena)) {
    return false;
  }
  const int lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(dst);
  const int src_lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(src);
  UPB_ASSERT(lg2 == src_lg2);
  char* dst_data = upb_Array_MutableDataPtr(dst);
  const char* src_data = upb_Array_DataPtr(src);
  memcpy(dst_data + (dst_len << lg2), src_data, src_len << lg2);
  return true;
}

void upb_Array_Move(upb_Array* arr, size_t dst_idx, size_t src_idx,
                    size_t count) {
  UPB_ASSERT(!upb_Array_IsFrozen(arr));
  const int lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(arr);
  char* data = upb_Array_MutableDataPtr(arr);
  memmove(&data[dst_idx << lg2], &data[src_idx << lg2], count << lg2);
}

bool upb_Array_Insert(upb_Array* arr, size_t i, size_t count,
                      upb_Arena* arena) {
  UPB_ASSERT(!upb_Array_IsFrozen(arr));
  UPB_ASSERT(arena);
  UPB_ASSERT(i <= arr->UPB_PRIVATE(size));
  UPB_ASSERT(count + arr->UPB_PRIVATE(size) >= count);
  const size_t oldsize = arr->UPB_PRIVATE(size);
  if (!UPB_PRIVATE(_upb_Array_ResizeUninitialized)(
          arr, arr->UPB_PRIVATE(size) + count, arena)) {
    return false;
  }
  upb_Array_Move(arr, i + count, i, oldsize - i);
  return true;
}

/*
 *              i        end      arr->size
 * |------------|XXXXXXXX|--------|
 */
void upb_Array_Delete(upb_Array* arr, size_t i, size_t count) {
  UPB_ASSERT(!upb_Array_IsFrozen(arr));
  const size_t end = i + count;
  UPB_ASSERT(i <= end);
  UPB_ASSERT(end <= arr->UPB_PRIVATE(size));
  upb_Array_Move(arr, i, end, arr->UPB_PRIVATE(size) - end);
  arr->UPB_PRIVATE(size) -= count;
}

bool upb_Array_Resize(upb_Array* arr, size_t size, upb_Arena* arena) {
  UPB_ASSERT(!upb_Array_IsFrozen(arr));
  const size_t oldsize = arr->UPB_PRIVATE(size);
  if (UPB_UNLIKELY(
          !UPB_PRIVATE(_upb_Array_ResizeUninitialized)(arr, size, arena))) {
    return false;
  }
  const size_t newsize = arr->UPB_PRIVATE(size);
  if (newsize > oldsize) {
    const int lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(arr);
    char* data = upb_Array_MutableDataPtr(arr);
    memset(data + (oldsize << lg2), 0, (newsize - oldsize) << lg2);
  }
  return true;
}

bool UPB_PRIVATE(_upb_Array_Realloc)(upb_Array* array, size_t min_capacity,
                                     upb_Arena* arena) {
  size_t new_capacity = UPB_MAX(array->UPB_PRIVATE(capacity), 4);
  const int lg2 = UPB_PRIVATE(_upb_Array_ElemSizeLg2)(array);
  size_t old_bytes = array->UPB_PRIVATE(capacity) << lg2;
  void* ptr = upb_Array_MutableDataPtr(array);

  // Log2 ceiling of size.
  while (new_capacity < min_capacity) {
    if (upb_ShlOverflow(&new_capacity, 1)) {
      new_capacity = SIZE_MAX;
      break;
    }
  }

  // If capacity doubling overflowed to SIZE_MAX, fail. No valid array can hold
  // SIZE_MAX elements, and downstream size calculations would overflow.
  if (new_capacity == SIZE_MAX) return false;

  size_t new_bytes = new_capacity;
  if (upb_ShlOverflow(&new_bytes, lg2)) {
    return false;
  }
  ptr = upb_Arena_Realloc(arena, ptr, old_bytes, new_bytes);
  if (!ptr) return false;

  UPB_PRIVATE(_upb_Array_SetTaggedPtr)(array, ptr, lg2);
  array->UPB_PRIVATE(capacity) = new_capacity;
  return true;
}

void upb_Array_Freeze(upb_Array* arr, const upb_MiniTable* m) {
  if (upb_Array_IsFrozen(arr)) return;
  UPB_PRIVATE(_upb_Array_ShallowFreeze)(arr);

  if (m) {
    const size_t size = upb_Array_Size(arr);

    for (size_t i = 0; i < size; i++) {
      upb_MessageValue val = upb_Array_Get(arr, i);
      upb_Message_Freeze((upb_Message*)val.msg_val, m);
    }
  }
}


#include <stdint.h>
#include <string.h>


// Must be last.

enum {
  kUpb_ConvertStatus_Ok = kUpb_ErrorCode_Ok,
  // The source and destination MiniTables are not compatible for conversion.
  kUpb_ConvertStatus_Incompatible = 10,
};

typedef struct {
  upb_Decoder decoder;
  upb_encstate encoder;
  upb_Arena* arena;
  upb_ErrorHandler err;
  int decode_options;
  int encode_options;
} upb_Converter;

// Minitable compatibility type check on the field, but not the
// submessage. Note: this check always succeeds for enums, whether the
// enum is open or closed.
UPB_INLINE bool _upb_MiniTableField_IsCompatible(
    const upb_MiniTableField* src_f, const upb_MiniTableField* dst_f) {
  return upb_MiniTableField_Type(src_f) == upb_MiniTableField_Type(dst_f) &&
         upb_MiniTableField_IsArray(src_f) ==
             upb_MiniTableField_IsArray(dst_f) &&
         upb_MiniTableField_IsMap(src_f) == upb_MiniTableField_IsMap(dst_f);
}

UPB_INLINE bool _upb_MiniTableField_IsMapEntryCompatible(
    const upb_MiniTableField* src_f, const upb_MiniTableField* dst_f) {
  const upb_MiniTable* src_entry_mt = upb_MiniTable_MapEntrySubMessage(src_f);
  const upb_MiniTable* dst_entry_mt = upb_MiniTable_MapEntrySubMessage(dst_f);
  if (src_entry_mt == dst_entry_mt) return true;
  return _upb_MiniTableField_IsCompatible(upb_MiniTable_MapKey(src_entry_mt),
                                          upb_MiniTable_MapKey(dst_entry_mt)) &&
         _upb_MiniTableField_IsCompatible(upb_MiniTable_MapValue(src_entry_mt),
                                          upb_MiniTable_MapValue(dst_entry_mt));
}

UPB_INLINE bool _upb_MiniTableField_IsExtensionCompatible(
    const upb_MiniTableField* src_f, const upb_MiniTableField* dst_f) {
  UPB_ASSERT(!upb_MiniTableField_IsMap(src_f));
  if (upb_MiniTableField_IsMap(dst_f)) return false;
  return upb_MiniTableField_Type(dst_f) == upb_MiniTableField_Type(src_f) &&
         upb_MiniTableField_IsArray(dst_f) == upb_MiniTableField_IsArray(src_f);
}

static void upb_Message_SetFieldOrExtension(upb_Message* msg,
                                            const upb_MiniTableField* f,
                                            const upb_MiniTableExtension* ext,
                                            const upb_MessageValue* val,
                                            upb_Arena* arena) {
  if (ext != NULL) {
    upb_Message_SetExtension(msg, ext, val, arena);
  } else {
    upb_Message_SetBaseField(msg, f, val);
  }
}

static void upb_Message_EncodeFieldAsUnknown(
    upb_encstate* e, upb_Message* dst, const upb_Message* src,
    const upb_MiniTableField* src_field, int depth, int options,
    upb_ErrorHandler* err) {
  size_t size;
  int encode_options = upb_Encode_LimitDepth(options, depth);
  char* buf = upb_BackAlloc_Init(&e->alloc, e->alloc.arena);
  UPB_PRIVATE(_upb_Encode_Field)(e, src, src_field, &buf, &size,
                                 encode_options);
  if (size > 0) {
    if (!UPB_PRIVATE(_upb_Message_AddUnknown)(dst, buf, size, e->alloc.arena,
                                              kUpb_AddUnknown_Alias)) {
      upb_ErrorHandler_ThrowError(err, kUpb_ErrorCode_OutOfMemory);
    }
  }
}

static void upb_Message_EncodeExtensionAsUnknown(
    upb_encstate* e, upb_Message* dst, const upb_MiniTable* dst_mt,
    const upb_MiniTableExtension* ext, upb_MessageValue val, int depth,
    int options, upb_ErrorHandler* err) {
  size_t size;
  int encode_options = upb_Encode_LimitDepth(options, depth);
  bool is_message_set = upb_MiniTable_IsMessageSet(dst_mt);
  char* buf = upb_BackAlloc_Init(&e->alloc, e->alloc.arena);
  UPB_PRIVATE(_upb_Encode_Extension)(e, ext, val, is_message_set, &buf, &size,
                                     encode_options);
  if (size > 0) {
    if (!UPB_PRIVATE(_upb_Message_AddUnknown)(dst, buf, size, e->alloc.arena,
                                              kUpb_AddUnknown_Alias)) {
      upb_ErrorHandler_ThrowError(err, kUpb_ErrorCode_OutOfMemory);
    }
  }
}

static void upb_Message_ConvertInternal(upb_Converter* c, upb_Message* dst,
                                        const upb_Message* src,
                                        const upb_MiniTable* dst_mt,
                                        const upb_MiniTable* src_mt,
                                        const upb_ExtensionRegistry* extreg,
                                        int depth);

static void upb_Array_DeepConvert(
    upb_Converter* c, upb_Array* dst, const upb_Array* src,
    const upb_MiniTable* dst_sub_mt, const upb_MiniTable* src_sub_mt,
    const upb_MiniTableField* dst_f, upb_Message* dst_msg,
    const upb_ExtensionRegistry* extreg, int depth) {
  size_t size = upb_Array_Size(src);
  if (!upb_Array_Resize(dst, size, c->arena)) {
    upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
  }
  size_t dst_i = 0;
  for (size_t i = 0; i < size; ++i) {
    upb_MessageValue src_val = upb_Array_Get(src, i);
    if (upb_MiniTableField_IsClosedEnum(dst_f)) {
      const upb_MiniTableEnum* dst_e = upb_MiniTable_GetSubEnumTable(dst_f);
      if (upb_MiniTableEnum_CheckValue(dst_e, src_val.int32_val)) {
        upb_MessageValue dst_val;
        dst_val.int32_val = src_val.int32_val;
        upb_Array_Set(dst, dst_i++, dst_val);
      } else if (!_upb_Encoder_AddEnumValueToUnknown(
                     dst_msg, dst_f, src_val.int32_val, c->arena)) {
        upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
      }
    } else if (dst_sub_mt) {
      const upb_Message* src_msg = src_val.msg_val;
      upb_Message* dst_sub = upb_Message_New(dst_sub_mt, c->arena);
      if (!dst_sub) {
        upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
      }
      upb_Message_ConvertInternal(c, dst_sub, src_msg, dst_sub_mt, src_sub_mt,
                                  extreg, depth);
      upb_MessageValue dst_val;
      dst_val.msg_val = dst_sub;
      upb_Array_Set(dst, dst_i++, dst_val);
    } else {
      // Open enum or primitive case.
      upb_Array_Set(dst, dst_i++, src_val);
    }
  }
  if (dst_i != size) {
    upb_Array_Resize(dst, dst_i, c->arena);
  }
}

static bool upb_Message_ConvertArrayField(upb_Converter* c, upb_Message* dst,
                                          const upb_Message* src,
                                          const upb_MiniTableField* dst_f,
                                          const upb_MiniTableField* src_f,
                                          const upb_ExtensionRegistry* extreg,
                                          int depth) {
  const upb_Array* src_arr = upb_Message_GetArray(src, src_f);
  if (!src_arr) return true;

  const upb_MiniTable* dst_sub_mt = upb_MiniTable_SubMessage(dst_f);
  const upb_MiniTable* src_sub_mt = upb_MiniTable_SubMessage(src_f);

  if (dst_sub_mt != src_sub_mt || upb_MiniTableField_IsClosedEnum(dst_f) ||
      upb_MiniTableField_IsClosedEnum(src_f)) {
    upb_Array* dst_arr =
        upb_Array_New(c->arena, upb_MiniTableField_CType(dst_f));
    if (!dst_arr)
      upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
    upb_Array_DeepConvert(c, dst_arr, src_arr, dst_sub_mt, src_sub_mt, dst_f,
                          dst, extreg, depth);
    upb_Message_SetBaseField(dst, dst_f, &dst_arr);
    return true;
  }
  // Fall through to a shallow copy.
  return false;
}

static void upb_Map_DeepConvert(
    upb_Converter* c, upb_Map* dst, const upb_Map* src,
    const upb_MiniTable* dst_entry_mt, const upb_MiniTable* src_entry_mt,
    const upb_MiniTableField* dst_map_f, upb_Message* dst_msg,
    const upb_ExtensionRegistry* extreg, int depth) {
  const upb_MiniTableField* dst_val_f = upb_MiniTable_MapValue(dst_entry_mt);
  const upb_MiniTable* dst_val_mt = upb_MiniTable_SubMessage(dst_val_f);
  const upb_MiniTableField* src_val_f = upb_MiniTable_MapValue(src_entry_mt);
  const upb_MiniTable* src_val_mt = upb_MiniTable_SubMessage(src_val_f);

  size_t iter = kUpb_Map_Begin;
  upb_MessageValue key, src_val;
  while (upb_Map_Next(src, &key, &src_val, &iter)) {
    if (dst_val_mt && src_val_mt) {
      const upb_Message* src_msg = src_val.msg_val;
      upb_Message* dst_sub = upb_Message_New(dst_val_mt, c->arena);
      if (!dst_sub) {
        upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
      }
      upb_Message_ConvertInternal(c, dst_sub, src_msg, dst_val_mt, src_val_mt,
                                  extreg, depth);
      upb_MessageValue dst_val;
      dst_val.msg_val = dst_sub;
      if (!upb_Map_Set(dst, key, dst_val, c->arena)) {
        upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
      }
    } else {
      // Scalar value.
      if (upb_MiniTableField_IsClosedEnum(dst_val_f)) {
        const upb_MiniTableEnum* dst_e =
            upb_MiniTable_GetSubEnumTable(dst_val_f);
        if (upb_MiniTableEnum_CheckValue(dst_e, src_val.int32_val)) {
          if (!upb_Map_Set(dst, key, src_val, c->arena)) {
            upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
          }
        } else {
          upb_Message* ent_msg = upb_Message_New(src_entry_mt, c->arena);
          if (!ent_msg) {
            upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
          }
          upb_Message_SetBaseField(ent_msg, upb_MiniTable_MapKey(src_entry_mt),
                                   &key);
          upb_Message_SetBaseField(
              ent_msg, upb_MiniTable_MapValue(src_entry_mt), &src_val);
          _upb_Encoder_AddMapEntryUnknown(dst_msg, dst_map_f, ent_msg,
                                          src_entry_mt, c->arena);
        }
      } else {
        if (!upb_Map_Set(dst, key, src_val, c->arena)) {
          upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
        }
      }
    }
  }
}

static bool upb_Message_ConvertMapField(upb_Converter* c, upb_Message* dst,
                                        const upb_Message* src,
                                        const upb_MiniTableField* dst_f,
                                        const upb_MiniTableField* src_f,
                                        const upb_ExtensionRegistry* extreg,
                                        int depth) {
  const upb_Map* src_map = upb_Message_GetMap(src, src_f);
  if (!src_map) return true;

  const upb_MiniTable* dst_entry_mt = upb_MiniTable_MapEntrySubMessage(dst_f);
  const upb_MiniTable* src_entry_mt = upb_MiniTable_MapEntrySubMessage(src_f);

  if (dst_entry_mt != src_entry_mt) {
    const upb_MiniTableField* dst_val_f = upb_MiniTable_MapValue(dst_entry_mt);
    upb_Map* dst_map = upb_Map_New(
        c->arena, upb_MiniTableField_CType(upb_MiniTable_MapKey(dst_entry_mt)),
        upb_MiniTableField_CType(dst_val_f));
    if (!dst_map) {
      upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
    }
    upb_Map_DeepConvert(c, dst_map, src_map, dst_entry_mt, src_entry_mt, dst_f,
                        dst, extreg, depth);
    upb_Message_SetBaseField(dst, dst_f, &dst_map);
    return true;
  }
  return false;
}

UPB_INLINE bool upb_Converter_NeedsClosedEnumDeepConvert(
    const upb_MiniTableField* dst_f, const upb_MiniTableField* src_f) {
  // Determines if an enum conversion requires deep conversion based on the
  // following combinations of closed/open src_f and dst_f:
  // 1. Open -> Open: Shallow copy. Never needs deep conversion (behaves like
  // primitives).
  // 2. Open -> Closed: Always needs deep conversion to validate values and
  //    move invalid values to unknowns.
  // 3. Closed -> Open: Needs deep conversion/copy for arrays to avoid mutating
  // the source array when unknowns are decoded into the open destination.
  // 4. Closed -> Closed: Only needs deep conversion if the target and source
  //    schemas differ. If the schemas match, we can safely shallow copy.
  if (upb_MiniTableField_IsClosedEnum(dst_f)) {
    return !upb_MiniTableField_IsClosedEnum(src_f) ||
           upb_MiniTable_GetSubEnumTable(dst_f) !=
               upb_MiniTable_GetSubEnumTable(src_f);
  }
  return upb_MiniTableField_IsClosedEnum(src_f) &&
         upb_MiniTableField_IsArray(src_f);
}

static void upb_Message_ConvertField(upb_Converter* c, upb_Message* dst,
                                     const upb_Message* src,
                                     const upb_MiniTableField* dst_f,
                                     const upb_MiniTableField* src_f,
                                     const upb_ExtensionRegistry* extreg,
                                     int depth) {
  if (upb_MiniTableField_HasPresence(src_f)) {
    if (!upb_Message_HasBaseField(src, src_f)) return;
  } else if (upb_MiniTableField_IsScalar(src_f)) {
    // For proto3 implicit scalar fields, we only need to copy if the source
    // field is set.
    const void* src_data = UPB_PRIVATE(_upb_Message_DataPtr)(src, src_f);
    if (UPB_PRIVATE(_upb_MiniTableField_DataIsZero)(src_f, src_data)) return;
  }

  if (upb_MiniTableField_CType(dst_f) == kUpb_CType_Message) {
    if (upb_MiniTableField_IsScalar(dst_f)) {
      const upb_Message* src_sub = upb_Message_GetMessage(src, src_f);
      if (!src_sub) return;

      const upb_MiniTable* dst_sub_mt = upb_MiniTable_SubMessage(dst_f);
      const upb_MiniTable* src_sub_mt = upb_MiniTable_SubMessage(src_f);

      if (dst_sub_mt == src_sub_mt) {
        upb_Message_SetMessage(dst, dst_f, (upb_Message*)src_sub);
        return;
      }

      upb_Message* dst_sub =
          upb_Message_GetOrCreateMutableMessage(dst, dst_f, c->arena);
      if (!dst_sub)
        upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
      upb_Message_ConvertInternal(c, dst_sub, src_sub, dst_sub_mt, src_sub_mt,
                                  extreg, depth);
      return;
    } else if (upb_MiniTableField_IsArray(dst_f)) {
      if (upb_Message_ConvertArrayField(c, dst, src, dst_f, src_f, extreg,
                                        depth)) {
        return;
      }
    } else if (upb_MiniTableField_IsMap(dst_f)) {
      if (UPB_UNLIKELY(
              !_upb_MiniTableField_IsMapEntryCompatible(src_f, dst_f))) {
        upb_ErrorHandler_ThrowError(&c->err, kUpb_ConvertStatus_Incompatible);
      }
      if (upb_Message_ConvertMapField(c, dst, src, dst_f, src_f, extreg,
                                      depth)) {
        return;
      }
    }
  } else if (upb_Converter_NeedsClosedEnumDeepConvert(dst_f, src_f)) {
    if (upb_MiniTableField_IsArray(dst_f)) {
      if (upb_Message_ConvertArrayField(c, dst, src, dst_f, src_f, extreg,
                                        depth)) {
        return;
      }
    } else if (upb_MiniTableField_IsMap(dst_f)) {
      if (upb_Message_ConvertMapField(c, dst, src, dst_f, src_f, extreg,
                                      depth)) {
        return;
      }
    } else {
      int32_t val;
      memcpy(&val, UPB_PRIVATE(_upb_Message_DataPtr)(src, src_f), 4);
      const upb_MiniTableEnum* dst_e = upb_MiniTable_GetSubEnumTable(dst_f);
      if (!upb_MiniTableEnum_CheckValue(dst_e, val)) {
        if (!_upb_Encoder_AddEnumValueToUnknown(dst, dst_f, val, c->arena)) {
          upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
        }
        return;
      }
    }
  }

  UPB_PRIVATE(_upb_MiniTableField_DataCopy)
  (dst_f, UPB_PRIVATE(_upb_Message_MutableDataPtr)(dst, dst_f),
   UPB_PRIVATE(_upb_Message_DataPtr)(src, src_f));

  if (upb_MiniTableField_HasPresence(dst_f)) {
    UPB_PRIVATE(_upb_Message_SetPresence)(dst, dst_f);
  }
}

static void upb_Message_ConvertExtensions(upb_Converter* c, upb_Message* dst,
                                          const upb_Message* src,
                                          const upb_MiniTable* dst_mt,
                                          const upb_ExtensionRegistry* extreg,
                                          int depth) {
  const upb_MiniTableExtension* ext;
  upb_MessageValue val;
  uintptr_t iter = kUpb_Message_ExtensionBegin;
  while (upb_Message_NextExtension(src, &ext, &val, &iter)) {
    const upb_MiniTableField* dst_f = upb_MiniTable_FindFieldByNumber(
        dst_mt, upb_MiniTableExtension_Number(ext));
    const upb_MiniTableExtension* dst_ext = NULL;
    if (!dst_f) {
      // Source extension not found in the destination schema. Check the
      // extension registry.
      if (extreg != NULL) {
        dst_ext = upb_ExtensionRegistry_Lookup(
            extreg, dst_mt, upb_MiniTableExtension_Number(ext));
        if (dst_ext) {
          dst_f = upb_MiniTableExtension_ToField(dst_ext);
        }
      }
    }

    if (dst_f) {
      const upb_MiniTableField* src_f = upb_MiniTableExtension_ToField(ext);

      UPB_ASSERT(!upb_MiniTableField_IsMap(src_f));
      if (UPB_UNLIKELY(
              !_upb_MiniTableField_IsExtensionCompatible(src_f, dst_f))) {
        // Return an error due to type mismatch.
        upb_ErrorHandler_ThrowError(&c->err, kUpb_ConvertStatus_Incompatible);
      }

      if (upb_MiniTableField_CType(dst_f) == kUpb_CType_Message) {
        const upb_MiniTable* dst_sub_mt = upb_MiniTable_SubMessage(dst_f);
        const upb_MiniTable* src_sub_mt = upb_MiniTable_SubMessage(src_f);

        if (upb_MiniTableField_IsArray(dst_f)) {
          if (dst_sub_mt != src_sub_mt) {
            // Array of messages, and the sub message types differ. Perform
            // conversion.
            upb_Array* dst_arr = upb_Array_New(c->arena, kUpb_CType_Message);
            if (!dst_arr)
              upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
            upb_Array_DeepConvert(c, dst_arr, val.array_val, dst_sub_mt,
                                  src_sub_mt, dst_f, dst, extreg, depth);
            upb_MessageValue valid_val;
            valid_val.array_val = dst_arr;
            upb_Message_SetFieldOrExtension(dst, dst_f, dst_ext, &valid_val,
                                            c->arena);
          } else {
            // Array of messages, and the sub message types are the same.
            // Shallow copy.
            upb_Message_SetFieldOrExtension(dst, dst_f, dst_ext, &val,
                                            c->arena);
          }
        } else if (dst_sub_mt == src_sub_mt) {
          // Scalar message, and the message types are the same.
          // Shallow copy.
          upb_Message_SetFieldOrExtension(dst, dst_f, dst_ext, &val, c->arena);
        } else {
          // Scalar message, and the message types differ. Perform conversion.
          upb_Message* dst_sub = upb_Message_New(dst_sub_mt, c->arena);
          if (!dst_sub)
            upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);

          upb_Message_ConvertInternal(c, dst_sub, val.msg_val, dst_sub_mt,
                                      src_sub_mt, extreg, depth);

          upb_MessageValue valid_val;
          valid_val.msg_val = dst_sub;
          upb_Message_SetFieldOrExtension(dst, dst_f, dst_ext, &valid_val,
                                          c->arena);
        }
      } else {
        // Scalar non-message type.
        if (upb_MiniTableField_IsClosedEnum(dst_f)) {
          if (upb_MiniTableField_IsArray(dst_f)) {
            upb_Array* dst_arr = upb_Array_New(c->arena, kUpb_CType_Int32);
            if (!dst_arr)
              upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_OutOfMemory);
            upb_Array_DeepConvert(c, dst_arr, val.array_val, NULL, NULL, dst_f,
                                  dst, extreg, depth);
            upb_MessageValue valid_val;
            valid_val.array_val = dst_arr;
            upb_Message_SetFieldOrExtension(dst, dst_f, dst_ext, &valid_val,
                                            c->arena);
            continue;
          } else {
            const upb_MiniTableEnum* dst_e =
                dst_ext ? upb_MiniTableExtension_GetSubEnum(dst_ext)
                        : upb_MiniTable_GetSubEnumTable(dst_f);
            if (!upb_MiniTableEnum_CheckValue(dst_e, val.int32_val)) {
              if (!_upb_Encoder_AddEnumValueToUnknown(dst, dst_f, val.int32_val,
                                                      c->arena)) {
                upb_ErrorHandler_ThrowError(&c->err,
                                            kUpb_ErrorCode_OutOfMemory);
              }
              continue;
            }
          }
        }
        upb_Message_SetFieldOrExtension(dst, dst_f, dst_ext, &val, c->arena);
      }
    } else {
      // Since this extension is not known in the destination schema, encode it
      // as an unknown field.
      // TODO - b/510055656: to handle this as a non-canonical extension
      upb_Message_EncodeExtensionAsUnknown(&c->encoder, dst, dst_mt, ext, val,
                                           depth, c->encode_options, &c->err);
    }
  }
}

static void upb_Message_ConvertInternal(upb_Converter* c, upb_Message* dst,
                                        const upb_Message* src,
                                        const upb_MiniTable* dst_mt,
                                        const upb_MiniTable* src_mt,
                                        const upb_ExtensionRegistry* extreg,
                                        int depth) {
  UPB_ASSERT(dst != NULL);
  if (--depth == 0) {
    upb_ErrorHandler_ThrowError(&c->err, kUpb_ErrorCode_MaxDepthExceeded);
  }

  // Bails out if the source and destination are not both MessageSets.
  if (upb_MiniTable_IsMessageSet(dst_mt) !=
      upb_MiniTable_IsMessageSet(src_mt)) {
    upb_ErrorHandler_ThrowError(&c->err, kUpb_ConvertStatus_Incompatible);
  }

  const upb_MiniTableField* dst_f = NULL;
  const upb_MiniTableField* dst_first = NULL;
  const upb_MiniTableField* src_f = NULL;
  const upb_MiniTableField* src_first = NULL;

  if (upb_MiniTable_FieldCount(dst_mt) > 0) {
    dst_first = upb_MiniTable_GetFieldByIndex(dst_mt, 0);
    dst_f = dst_first + upb_MiniTable_FieldCount(dst_mt);
  }
  if (upb_MiniTable_FieldCount(src_mt) > 0) {
    src_first = upb_MiniTable_GetFieldByIndex(src_mt, 0);
    src_f = src_first + upb_MiniTable_FieldCount(src_mt);
  }

  // Convert fields in descending order of field number.
  while (dst_f != dst_first || src_f != src_first) {
    uint32_t dst_nr =
        dst_f != dst_first ? upb_MiniTableField_Number(dst_f - 1) : 0;
    uint32_t src_nr =
        src_f != src_first ? upb_MiniTableField_Number(src_f - 1) : 0;

    if (dst_nr == src_nr) {
      const upb_MiniTableField* dst_next = dst_f - 1;
      const upb_MiniTableField* src_next = src_f - 1;

      if (UPB_UNLIKELY(!_upb_MiniTableField_IsCompatible(src_next, dst_next))) {
        upb_ErrorHandler_ThrowError(&c->err, kUpb_ConvertStatus_Incompatible);
      }
      if (upb_MiniTableField_IsInOneof(dst_next) &&
          UPB_PRIVATE(_upb_Message_GetOneofCase)(dst, dst_next) != 0) {
        // Since fields are processed in descending order, the first encountered
        // oneof field is the one that wins. We ignore subsequent ones to match
        // the encoding-then-decoding behavior.
      } else {
        upb_Message_ConvertField(c, dst, src, dst_next, src_next, extreg,
                                 depth);
      }
      dst_f--;
      src_f--;
    } else if (dst_nr > src_nr) {
      dst_f--;
    } else {
      const upb_MiniTableField* src_next = src_f - 1;
      upb_Message_EncodeFieldAsUnknown(&c->encoder, dst, src, src_next, depth,
                                       c->encode_options, &c->err);
      src_f--;
    }
  }

  // Convert extensions.
  if (src_mt->UPB_PRIVATE(ext) != kUpb_ExtMode_NonExtendable) {
    upb_Message_ConvertExtensions(c, dst, src, dst_mt, extreg, depth);
  }

  // Convert unknown fields.
  upb_StringView data;
  size_t iter = kUpb_Message_UnknownBegin;
  while (upb_Message_NextUnknown(src, &data, &iter)) {
    int decode_options = upb_Decode_LimitDepth(
        c->decode_options | kUpb_DecodeOption_AliasString, depth);

    // Reuse d. Reset input stream.
    const char* ptr = data.data;
    upb_Decoder* d = &c->decoder;
    upb_EpsCopyInputStream_InitWithErrorHandler(&d->input, &ptr, data.size,
                                                d->err);
    upb_Decoder_Reset(d, decode_options, dst);
    _upb_Decoder_DecodeMessage(d, ptr, dst, dst_mt);
    UPB_ASSERT(d->end_group == DECODE_NOGROUP);
  }
}

static bool upb_Message_DoConvert(upb_Converter* c, upb_Message* dst,
                                  const upb_Message* src,
                                  const upb_MiniTable* dst_mt,
                                  const upb_MiniTable* src_mt,
                                  const upb_ExtensionRegistry* extreg) {
  if (UPB_SETJMP(c->err.buf) == 0) {
    upb_Message_ConvertInternal(c, dst, src, dst_mt, src_mt, extreg, 100);
    return true;
  }
  return false;
}

const upb_Message* upb_Message_Convert(const upb_Message* src,
                                       const upb_MiniTable* src_mt,
                                       const upb_MiniTable* dst_mt,
                                       const upb_ExtensionRegistry* extreg,
                                       int decode_options, int encode_options,
                                       upb_Arena* arena) {
  if (dst_mt == src_mt && extreg == NULL) return src;

  upb_Message* dst = upb_Message_New(dst_mt, arena);
  if (!dst) return NULL;

  upb_Converter c;
  upb_ErrorHandler_Init(&c.err);
  c.decode_options = decode_options;
  c.encode_options = encode_options;

  // Initialize the decoder.
  // Initialize decoder once, performing SwapIn.
  // We use a NULL buffer initially, effectively a dummy init to set up the
  // arena and error handler. Note: we pass &c.err.
  upb_Decoder_Init(&c.decoder, NULL, 0, extreg, decode_options, arena, &c.err,
                   NULL, 0);

  // Initialize the encoder.
  UPB_PRIVATE(_upb_encstate_init)(&c.encoder, &c.err.buf, &c.decoder.arena);

  c.arena = &c.decoder.arena;

  if (!upb_Message_DoConvert(&c, dst, src, dst_mt, src_mt, extreg)) {
    dst = NULL;
  }

#ifndef NDEBUG
  if (dst) {
    char* wire_buf;
    size_t wire_size;
    upb_Arena* tmp_arena = upb_Arena_New();

    // Compare the encoded/decoded round-trip of the original message to the
    // converted message.
    // Encode/decode original message `src`
    upb_EncodeStatus encode_status = upb_Encode(
        src, src_mt, encode_options, tmp_arena, &wire_buf, &wire_size);
    if (encode_status != kUpb_EncodeStatus_MaxDepthExceeded) {
      UPB_ASSERT(encode_status == kUpb_EncodeStatus_Ok);
      upb_Message* decoded_msg = upb_Message_New(dst_mt, tmp_arena);
      upb_DecodeStatus decode_status =
          upb_Decode(wire_buf, wire_size, decoded_msg, dst_mt, extreg,
                     decode_options, tmp_arena);
      if (decode_status != kUpb_DecodeStatus_MaxDepthExceeded) {
        UPB_ASSERT(decode_status == kUpb_DecodeStatus_Ok);
        // Compare the decoded message to the converted message.
        UPB_ASSERT(upb_Message_IsEqual(decoded_msg, dst, dst_mt, 0));
      }
    }
    upb_Arena_Free(tmp_arena);
  }
#endif

  upb_Decoder_Destroy(&c.decoder, arena);
  UPB_PRIVATE(_upb_encstate_destroy)(&c.encoder);
  return dst;
}


#include <stdint.h>
#include <stdlib.h>


// Must be last.

typedef struct upb_UnknownFields upb_UnknownFields;

typedef struct {
  uint32_t tag;
  union {
    uint64_t varint;
    uint64_t uint64;
    uint32_t uint32;
    upb_StringView delimited;
    upb_UnknownFields* group;
  } data;
} upb_UnknownField;

struct upb_UnknownFields {
  size_t size;
  size_t capacity;
  upb_UnknownField* fields;
};

typedef struct {
  upb_EpsCopyInputStream stream;
  upb_Arena* arena;
  upb_UnknownField* tmp;
  size_t tmp_size;
  int depth;
  upb_UnknownCompareResult status;
  jmp_buf err;
} upb_UnknownField_Context;

typedef struct {
  upb_UnknownField* arr_base;
  upb_UnknownField* arr_ptr;
  upb_UnknownField* arr_end;
  uint32_t last_tag;
  bool sorted;
} upb_UnknownFields_Builder;

UPB_NORETURN static void upb_UnknownFields_OutOfMemory(
    upb_UnknownField_Context* ctx) {
  ctx->status = kUpb_UnknownCompareResult_OutOfMemory;
  UPB_LONGJMP(ctx->err, 1);
}

static void upb_UnknownFields_Grow(upb_UnknownField_Context* ctx,
                                   upb_UnknownField** base,
                                   upb_UnknownField** ptr,
                                   upb_UnknownField** end) {
  size_t old = (*ptr - *base);
  size_t new = UPB_MAX(4, old * 2);

  *base = upb_Arena_Realloc(ctx->arena, *base, old * sizeof(**base),
                            new * sizeof(**base));
  if (!*base) upb_UnknownFields_OutOfMemory(ctx);

  *ptr = *base + old;
  *end = *base + new;
}

// We have to implement our own sort here, since qsort() is not an in-order
// sort. Here we use merge sort, the simplest in-order sort.
static void upb_UnknownFields_Merge(upb_UnknownField* arr, size_t start,
                                    size_t mid, size_t end,
                                    upb_UnknownField* tmp) {
  memcpy(tmp, &arr[start], (end - start) * sizeof(*tmp));

  upb_UnknownField* ptr1 = tmp;
  upb_UnknownField* end1 = &tmp[mid - start];
  upb_UnknownField* ptr2 = &tmp[mid - start];
  upb_UnknownField* end2 = &tmp[end - start];
  upb_UnknownField* out = &arr[start];

  while (ptr1 < end1 && ptr2 < end2) {
    if (ptr1->tag <= ptr2->tag) {
      *out++ = *ptr1++;
    } else {
      *out++ = *ptr2++;
    }
  }

  if (ptr1 < end1) {
    memcpy(out, ptr1, (end1 - ptr1) * sizeof(*out));
  } else if (ptr2 < end2) {
    memcpy(out, ptr2, (end2 - ptr2) * sizeof(*out));
  }
}

static void upb_UnknownFields_SortRecursive(upb_UnknownField* arr, size_t start,
                                            size_t end, upb_UnknownField* tmp) {
  if (end - start > 1) {
    size_t mid = start + ((end - start) / 2);
    upb_UnknownFields_SortRecursive(arr, start, mid, tmp);
    upb_UnknownFields_SortRecursive(arr, mid, end, tmp);
    upb_UnknownFields_Merge(arr, start, mid, end, tmp);
  }
}

static void upb_UnknownFields_Sort(upb_UnknownField_Context* ctx,
                                   upb_UnknownFields* fields) {
  if (ctx->tmp_size < fields->size) {
    const int oldsize = ctx->tmp_size * sizeof(*ctx->tmp);
    ctx->tmp_size = UPB_MAX(8, ctx->tmp_size);
    while (ctx->tmp_size < fields->size) ctx->tmp_size *= 2;
    const int newsize = ctx->tmp_size * sizeof(*ctx->tmp);
    ctx->tmp = upb_grealloc(ctx->tmp, oldsize, newsize);
  }
  upb_UnknownFields_SortRecursive(fields->fields, 0, fields->size, ctx->tmp);
}

static upb_UnknownFields* upb_UnknownFields_BuildFromBuffer(
    upb_UnknownField_Context* ctx, const char** buf);

// Combines two unknown fields into one.
static void upb_CombineUnknownFields(upb_UnknownField_Context* ctx,
                                     upb_UnknownFields_Builder* builder,
                                     const char** buf) {
  upb_UnknownField* arr_base = builder->arr_base;
  upb_UnknownField* arr_ptr = builder->arr_ptr;
  upb_UnknownField* arr_end = builder->arr_end;
  const char* ptr = *buf;
  uint32_t last_tag = builder->last_tag;
  bool sorted = builder->sorted;

  // Parse the unknown field data. It is an invariant of the data structure that
  // unknown field data is valid, so parse errors here should be impossible.
  while (!upb_EpsCopyInputStream_IsDone(&ctx->stream, &ptr)) {
    uint32_t tag;
    ptr = upb_WireReader_ReadTag(ptr, &tag, &ctx->stream);
    UPB_ASSERT(tag <= UINT32_MAX);
    int wire_type = upb_WireReader_GetWireType(tag);
    if (wire_type == kUpb_WireType_EndGroup) break;
    if (tag < last_tag) sorted = false;
    last_tag = tag;

    if (arr_ptr == arr_end) {
      upb_UnknownFields_Grow(ctx, &arr_base, &arr_ptr, &arr_end);
    }
    upb_UnknownField* field = arr_ptr;
    field->tag = tag;
    arr_ptr++;

    switch (wire_type) {
      case kUpb_WireType_Varint:
        ptr = upb_WireReader_ReadVarint(ptr, &field->data.varint, &ctx->stream);
        UPB_ASSERT(ptr);
        break;
      case kUpb_WireType_64Bit:
        ptr =
            upb_WireReader_ReadFixed64(ptr, &field->data.uint64, &ctx->stream);
        UPB_ASSERT(ptr);
        break;
      case kUpb_WireType_32Bit:
        ptr =
            upb_WireReader_ReadFixed32(ptr, &field->data.uint32, &ctx->stream);
        UPB_ASSERT(ptr);
        break;
      case kUpb_WireType_Delimited: {
        int size;
        upb_StringView sv;
        ptr = upb_WireReader_ReadSize(ptr, &size, &ctx->stream);
        UPB_ASSERT(ptr);
        ptr = upb_EpsCopyInputStream_ReadStringAlwaysAlias(&ctx->stream, ptr,
                                                           size, &sv);
        UPB_ASSERT(ptr);
        field->data.delimited.data = sv.data;
        field->data.delimited.size = sv.size;
        break;
      }
      case kUpb_WireType_StartGroup:
        if (--ctx->depth < 0) {
          ctx->status = kUpb_UnknownCompareResult_MaxDepthExceeded;
          UPB_LONGJMP(ctx->err, 1);
        }
        field->data.group = upb_UnknownFields_BuildFromBuffer(ctx, &ptr);
        ctx->depth++;
        break;
      default:
        UPB_UNREACHABLE();
    }
  }
  *buf = ptr;
  builder->arr_base = arr_base;
  builder->arr_ptr = arr_ptr;
  builder->arr_end = arr_end;
  builder->sorted = sorted;
  builder->last_tag = last_tag;
}

static upb_UnknownFields* upb_UnknownFields_DoBuild(
    upb_UnknownField_Context* ctx, upb_UnknownFields_Builder* builder) {
  upb_UnknownFields* ret = upb_Arena_Malloc(ctx->arena, sizeof(*ret));
  if (!ret) upb_UnknownFields_OutOfMemory(ctx);
  ret->fields = builder->arr_base;
  ret->size = builder->arr_ptr - builder->arr_base;
  ret->capacity = builder->arr_end - builder->arr_base;
  if (!builder->sorted) {
    upb_UnknownFields_Sort(ctx, ret);
  }
  return ret;
}

// Builds a upb_UnknownFields data structure from the binary data in buf.
static upb_UnknownFields* upb_UnknownFields_BuildFromBuffer(
    upb_UnknownField_Context* ctx, const char** buf) {
  upb_UnknownFields_Builder builder = {
      .arr_base = NULL,
      .arr_ptr = NULL,
      .arr_end = NULL,
      .sorted = true,
      .last_tag = 0,
  };
  const char* ptr = *buf;
  upb_CombineUnknownFields(ctx, &builder, &ptr);
  upb_UnknownFields* fields = upb_UnknownFields_DoBuild(ctx, &builder);
  *buf = ptr;
  return fields;
}

// Builds a upb_UnknownFields data structure from the unknown fields of a
// upb_Message.
static upb_UnknownFields* upb_UnknownFields_Build(upb_UnknownField_Context* ctx,
                                                  const upb_Message* msg) {
  upb_UnknownFields_Builder builder = {
      .arr_base = NULL,
      .arr_ptr = NULL,
      .arr_end = NULL,
      .sorted = true,
      .last_tag = 0,
  };
  uintptr_t iter = kUpb_Message_UnknownBegin;
  upb_StringView view;
  while (upb_Message_NextUnknown(msg, &view, &iter)) {
    upb_EpsCopyInputStream_Init(&ctx->stream, &view.data, view.size);
    upb_CombineUnknownFields(ctx, &builder, &view.data);
    UPB_ASSERT(upb_EpsCopyInputStream_IsDone(&ctx->stream, &view.data) &&
               !upb_EpsCopyInputStream_IsError(&ctx->stream));
  }
  upb_UnknownFields* fields = upb_UnknownFields_DoBuild(ctx, &builder);
  return fields;
}

// Compares two sorted upb_UnknownFields structures for equality.
static bool upb_UnknownFields_IsEqual(const upb_UnknownFields* uf1,
                                      const upb_UnknownFields* uf2) {
  if (uf1->size != uf2->size) return false;
  for (size_t i = 0, n = uf1->size; i < n; i++) {
    upb_UnknownField* f1 = &uf1->fields[i];
    upb_UnknownField* f2 = &uf2->fields[i];
    if (f1->tag != f2->tag) return false;
    int wire_type = f1->tag & 7;
    switch (wire_type) {
      case kUpb_WireType_Varint:
        if (f1->data.varint != f2->data.varint) return false;
        break;
      case kUpb_WireType_64Bit:
        if (f1->data.uint64 != f2->data.uint64) return false;
        break;
      case kUpb_WireType_32Bit:
        if (f1->data.uint32 != f2->data.uint32) return false;
        break;
      case kUpb_WireType_Delimited:
        if (!upb_StringView_IsEqual(f1->data.delimited, f2->data.delimited)) {
          return false;
        }
        break;
      case kUpb_WireType_StartGroup:
        if (!upb_UnknownFields_IsEqual(f1->data.group, f2->data.group)) {
          return false;
        }
        break;
      default:
        UPB_UNREACHABLE();
    }
  }
  return true;
}

static upb_UnknownCompareResult upb_UnknownField_DoCompare(
    upb_UnknownField_Context* ctx, const upb_Message* msg1,
    const upb_Message* msg2) {
  upb_UnknownCompareResult ret;
  // First build both unknown fields into a sorted data structure (similar
  // to the UnknownFieldSet in C++).
  upb_UnknownFields* uf1 = upb_UnknownFields_Build(ctx, msg1);
  upb_UnknownFields* uf2 = upb_UnknownFields_Build(ctx, msg2);

  // Now perform the equality check on the sorted structures.
  if (upb_UnknownFields_IsEqual(uf1, uf2)) {
    ret = kUpb_UnknownCompareResult_Equal;
  } else {
    ret = kUpb_UnknownCompareResult_NotEqual;
  }
  return ret;
}

static upb_UnknownCompareResult upb_UnknownField_Compare(
    upb_UnknownField_Context* const ctx, const upb_Message* msg1,
    const upb_Message* msg2) {
  upb_UnknownCompareResult ret;
  if (UPB_SETJMP(ctx->err) == 0) {
    ret = upb_UnknownField_DoCompare(ctx, msg1, msg2);
  } else {
    ret = ctx->status;
    UPB_ASSERT(ret != kUpb_UnknownCompareResult_Equal);
  }

  upb_Arena_Free(ctx->arena);
  upb_gfree(ctx->tmp);
  return ret;
}

upb_UnknownCompareResult UPB_PRIVATE(_upb_Message_UnknownFieldsAreEqual)(
    const upb_Message* msg1, const upb_Message* msg2, int max_depth) {
  bool msg1_empty = !upb_Message_HasUnknown(msg1);
  bool msg2_empty = !upb_Message_HasUnknown(msg2);
  if (msg1_empty && msg2_empty) return kUpb_UnknownCompareResult_Equal;
  if (msg1_empty || msg2_empty) return kUpb_UnknownCompareResult_NotEqual;

  upb_UnknownField_Context ctx = {
      .arena = upb_Arena_New(),
      .depth = max_depth,
      .tmp = NULL,
      .tmp_size = 0,
      .status = kUpb_UnknownCompareResult_Equal,
  };

  if (!ctx.arena) return kUpb_UnknownCompareResult_OutOfMemory;

  return upb_UnknownField_Compare(&ctx, msg1, msg2);
}


#include <stdint.h>
#include <string.h>


// Must be last.

const upb_Extension* UPB_PRIVATE(_upb_Message_Getext)(
    const struct upb_Message* msg, const upb_MiniTableExtension* e) {
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  if (!in) return NULL;

  for (size_t i = 0; i < in->size; i++) {
    upb_TaggedAuxPtr tagged_ptr = in->aux_data[i];
    if (upb_TaggedAuxPtr_IsCanonicalExtension(tagged_ptr)) {
      const upb_Extension* ext =
          upb_TaggedAuxPtr_CanonicalExtension(tagged_ptr);
      if (ext->ext == e) {
        return ext;
      }
    }
  }

  return NULL;
}

UPB_INLINE upb_Extension* _upb_Message_GetOrCreateExtensionInternal(
    struct upb_Message* msg, const upb_MiniTableExtension* e, upb_Arena* a,
    bool canonical) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  upb_Extension* ext = (upb_Extension*)UPB_PRIVATE(_upb_Message_Getext)(msg, e);
  if (ext) return ext;

  if (!UPB_PRIVATE(_upb_Message_ReserveSlot)(msg, a)) return NULL;
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  ext = upb_Arena_Malloc(a, sizeof(upb_Extension));
  if (!ext) return NULL;
  memset(ext, 0, sizeof(upb_Extension));
  ext->ext = e;
  in->aux_data[in->size++] =
      canonical ? upb_TaggedAuxPtr_MakeCanonicalExtension(ext)
                : upb_TaggedAuxPtr_MakeNonCanonicalExtension(ext);
  return ext;
}

upb_Extension* UPB_PRIVATE(_upb_Message_GetOrCreateExtension)(
    struct upb_Message* msg, const upb_MiniTableExtension* e, upb_Arena* a) {
  return _upb_Message_GetOrCreateExtensionInternal(msg, e, a,
                                                   /*canonical=*/true);
}

upb_Extension* UPB_PRIVATE(_upb_Message_CreateNonCanonicalExtension)(
    struct upb_Message* msg, const upb_MiniTableExtension* e, upb_Arena* a) {
  return _upb_Message_GetOrCreateExtensionInternal(msg, e, a,
                                                   /*canonical=*/false);
}


#include <stddef.h>


// Must be last.

bool UPB_PRIVATE(_upb_Message_NextBaseField)(const upb_Message* msg,
                                             const upb_MiniTable* m,
                                             const upb_MiniTableField** out_f,
                                             upb_MessageValue* out_v,
                                             size_t* iter) {
  const size_t count = upb_MiniTable_FieldCount(m);
  size_t i = *iter;

  while (++i < count) {
    const upb_MiniTableField* f = upb_MiniTable_GetFieldByIndex(m, i);
    const void* src = UPB_PRIVATE(_upb_Message_DataPtr)(msg, f);

    upb_MessageValue val;
    UPB_PRIVATE(_upb_MiniTableField_DataCopy)(f, &val, src);

    // Skip field if unset or empty.
    if (upb_MiniTableField_HasPresence(f)) {
      if (!upb_Message_HasBaseField(msg, f)) continue;
    } else {
      if (UPB_PRIVATE(_upb_MiniTableField_DataIsZero)(f, src)) continue;

      if (upb_MiniTableField_IsArray(f)) {
        if (upb_Array_Size(val.array_val) == 0) continue;
      } else if (upb_MiniTableField_IsMap(f)) {
        if (upb_Map_Size(val.map_val) == 0) continue;
      }
    }

    *out_f = f;
    *out_v = val;
    *iter = i;
    return true;
  }

  return false;
}


#include <math.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Must be last.

// The latest win32 SDKs have an invalid definition of NAN.
// https://developercommunity.visualstudio.com/t/NAN-is-no-longer-compile-time-constant-i/10688907
//
// Unfortunately, the `0.0 / 0.0` workaround doesn't work in Clang under C23, so
// try __builtin_nan first, if that exists.
#ifdef _WIN32
#ifdef __has_builtin
#if __has_builtin(__builtin_nan)
#define UPB_NAN __builtin_nan("0")
#endif
#if __has_builtin(__builtin_inf)
#define UPB_INFINITY __builtin_inf()
#endif
#endif
#ifndef UPB_NAN
#define UPB_NAN 0.0 / 0.0
#endif
#ifndef UPB_INFINITY
#define UPB_INFINITY 1.0 / 0.0
#endif
#else
// For !_WIN32, assume math.h works.
#define UPB_NAN NAN
#define UPB_INFINITY INFINITY
#endif

const float kUpb_FltInfinity = UPB_INFINITY;
const double kUpb_Infinity = UPB_INFINITY;
const double kUpb_NaN = UPB_NAN;

static size_t _upb_Message_SizeOfInternal(uint32_t count) {
  return UPB_SIZEOF_FLEX(upb_Message_Internal, aux_data, count);
}

bool UPB_PRIVATE(_upb_Message_ReserveSlot)(struct upb_Message* msg,
                                           upb_Arena* a) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  if (!in) {
    // No internal data, allocate from scratch.
    uint32_t capacity = 4;
    in = upb_Arena_Malloc(a, _upb_Message_SizeOfInternal(capacity));
    if (!in) return false;
    in->size = 0;
    in->capacity = capacity;
    UPB_PRIVATE(_upb_Message_SetInternal)(msg, in);
  } else if (in->capacity == in->size) {
    if (in->size == UINT32_MAX) return false;
    // Internal data is too small, reallocate.
    size_t needed_pow2 = upb_RoundUpToPowerOfTwo(in->size + 1);
    if (needed_pow2 > UINT32_MAX) return false;
    uint32_t new_capacity = needed_pow2;
    if (UPB_SIZEOF_FLEX_WOULD_OVERFLOW(upb_Message_Internal, aux_data,
                                       new_capacity)) {
      return false;
    }
    in = upb_Arena_Realloc(a, in, _upb_Message_SizeOfInternal(in->capacity),
                           _upb_Message_SizeOfInternal(new_capacity));
    if (!in) return false;
    in->capacity = new_capacity;
    UPB_PRIVATE(_upb_Message_SetInternal)(msg, in);
  }
  UPB_ASSERT(in->capacity - in->size >= 1);
  return true;
}

#ifdef UPB_TRACING_ENABLED
static void (*_message_trace_handler)(const upb_MiniTable*, const upb_Arena*);

void upb_Message_LogNewMessage(const upb_MiniTable* m, const upb_Arena* arena) {
  if (_message_trace_handler) {
    _message_trace_handler(m, arena);
  }
}

void upb_Message_SetNewMessageTraceHandler(void (*handler)(const upb_MiniTable*,
                                                           const upb_Arena*)) {
  _message_trace_handler = handler;
}
#endif  // UPB_TRACING_ENABLED


#include <stdbool.h>
#include <stdint.h>
#include <string.h>


// Must be last.

static upb_StringView upb_Clone_StringView(upb_StringView str,
                                           upb_Arena* arena) {
  if (str.size == 0) {
    return upb_StringView_FromDataAndSize(NULL, 0);
  }
  void* cloned_data = upb_Arena_Malloc(arena, str.size);
  upb_StringView cloned_str =
      upb_StringView_FromDataAndSize(cloned_data, str.size);
  memcpy(cloned_data, str.data, str.size);
  return cloned_str;
}

static bool upb_Clone_MessageValue(void* value, upb_CType value_type,
                                   const upb_MiniTable* sub, upb_Arena* arena) {
  switch (value_type) {
    case kUpb_CType_Bool:
    case kUpb_CType_Float:
    case kUpb_CType_Int32:
    case kUpb_CType_UInt32:
    case kUpb_CType_Enum:
    case kUpb_CType_Double:
    case kUpb_CType_Int64:
    case kUpb_CType_UInt64:
      return true;
    case kUpb_CType_String:
    case kUpb_CType_Bytes: {
      upb_StringView source = *(upb_StringView*)value;
      size_t size = source.size;
      void* cloned_data = upb_Arena_Malloc(arena, size);
      if (cloned_data == NULL) {
        return false;
      }
      *(upb_StringView*)value =
          upb_StringView_FromDataAndSize(cloned_data, size);
      memcpy(cloned_data, source.data, size);
      return true;
    } break;
    case kUpb_CType_Message: {
      UPB_ASSERT(sub);
      const upb_Message* source = *(upb_Message**)value;
      UPB_ASSERT(source);
      upb_Message* clone = upb_Message_DeepClone(source, sub, arena);
      *(upb_Message**)value = clone;
      return clone != NULL;
    } break;
  }
  UPB_UNREACHABLE();
}

upb_Map* upb_Map_DeepClone(const upb_Map* map, upb_CType key_type,
                           upb_CType value_type,
                           const upb_MiniTable* map_entry_table,
                           upb_Arena* arena) {
  upb_Map* cloned_map = _upb_Map_New(arena, map->key_size, map->val_size);
  if (cloned_map == NULL) {
    return NULL;
  }
  upb_MessageValue key, val;
  size_t iter = kUpb_Map_Begin;
  while (upb_Map_Next(map, &key, &val, &iter)) {
    const upb_MiniTableField* value_field =
        upb_MiniTable_MapValue(map_entry_table);
    const upb_MiniTable* value_sub =
        upb_MiniTableField_CType(value_field) == kUpb_CType_Message
            ? upb_MiniTable_GetSubMessageTable(value_field)
            : NULL;
    upb_CType value_field_type = upb_MiniTableField_CType(value_field);
    if (!upb_Clone_MessageValue(&val, value_field_type, value_sub, arena)) {
      return NULL;
    }
    if (!upb_Map_Set(cloned_map, key, val, arena)) {
      return NULL;
    }
  }
  return cloned_map;
}

static upb_Map* upb_Message_Map_DeepClone(const upb_Map* map,
                                          const upb_MiniTable* mini_table,
                                          const upb_MiniTableField* f,
                                          upb_Message* clone,
                                          upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(clone));
  const upb_MiniTable* map_entry_table = upb_MiniTable_MapEntrySubMessage(f);
  UPB_ASSERT(map_entry_table);

  const upb_MiniTableField* key_field = upb_MiniTable_MapKey(map_entry_table);
  const upb_MiniTableField* value_field =
      upb_MiniTable_MapValue(map_entry_table);

  upb_Map* cloned_map = upb_Map_DeepClone(
      map, upb_MiniTableField_CType(key_field),
      upb_MiniTableField_CType(value_field), map_entry_table, arena);
  if (!cloned_map) {
    return NULL;
  }
  upb_Message_SetBaseField(clone, f, &cloned_map);
  return cloned_map;
}

upb_Array* upb_Array_DeepClone(const upb_Array* array, upb_CType value_type,
                               const upb_MiniTable* sub, upb_Arena* arena) {
  const size_t size = upb_Array_Size(array);
  const int lg2 = UPB_PRIVATE(_upb_CType_SizeLg2)(value_type);
  upb_Array* cloned_array = UPB_PRIVATE(_upb_Array_New)(arena, size, lg2);
  if (!cloned_array) {
    return NULL;
  }
  if (!UPB_PRIVATE(_upb_Array_ResizeUninitialized)(cloned_array, size, arena)) {
    return NULL;
  }
  for (size_t i = 0; i < size; ++i) {
    upb_MessageValue val = upb_Array_Get(array, i);
    if (!upb_Clone_MessageValue(&val, value_type, sub, arena)) {
      return NULL;
    }
    upb_Array_Set(cloned_array, i, val);
  }
  return cloned_array;
}

static bool upb_Message_Array_DeepClone(const upb_Array* array,
                                        const upb_MiniTable* mini_table,
                                        const upb_MiniTableField* field,
                                        upb_Message* clone, upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(clone));
  UPB_PRIVATE(_upb_MiniTableField_CheckIsArray)(field);
  upb_Array* cloned_array =
      upb_Array_DeepClone(array, upb_MiniTableField_CType(field),
                          upb_MiniTableField_CType(field) == kUpb_CType_Message
                              ? upb_MiniTable_GetSubMessageTable(field)
                              : NULL,
                          arena);

  // Clear out upb_Array* due to parent memcpy.
  upb_Message_SetBaseField(clone, field, &cloned_array);
  return true;
}

static bool upb_Clone_ExtensionValue(
    const upb_MiniTableExtension* mini_table_ext, const upb_Extension* source,
    upb_Extension* dest, upb_Arena* arena) {
  dest->data = source->data;
  return upb_Clone_MessageValue(
      &dest->data, upb_MiniTableExtension_CType(mini_table_ext),
      upb_MiniTableExtension_GetSubMessage(mini_table_ext), arena);
}

upb_Message* _upb_Message_Copy(upb_Message* dst, const upb_Message* src,
                               const upb_MiniTable* mini_table,
                               upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(dst));
  upb_StringView empty_string = upb_StringView_FromDataAndSize(NULL, 0);
  // Only copy message area skipping upb_Message_Internal.
  memcpy(dst + 1, src + 1, mini_table->UPB_PRIVATE(size) - sizeof(upb_Message));
  for (int i = 0; i < upb_MiniTable_FieldCount(mini_table); ++i) {
    const upb_MiniTableField* field =
        upb_MiniTable_GetFieldByIndex(mini_table, i);
    if (upb_MiniTableField_IsScalar(field)) {
      switch (upb_MiniTableField_CType(field)) {
        case kUpb_CType_Message: {
          const upb_Message* sub_message = upb_Message_GetMessage(src, field);
          if (sub_message != NULL) {
            const upb_MiniTable* sub_message_table =
                upb_MiniTable_GetSubMessageTable(field);
            upb_Message* dst_sub_message =
                upb_Message_DeepClone(sub_message, sub_message_table, arena);
            if (dst_sub_message == NULL) {
              return NULL;
            }
            upb_Message_SetBaseFieldMessage(dst, field, dst_sub_message);
          }
        } break;
        case kUpb_CType_String:
        case kUpb_CType_Bytes: {
          upb_StringView str = upb_Message_GetString(src, field, empty_string);
          if (str.size != 0) {
            if (!upb_Message_SetString(
                    dst, field, upb_Clone_StringView(str, arena), arena)) {
              return NULL;
            }
          }
        } break;
        default:
          // Scalar, already copied.
          break;
      }
    } else {
      if (upb_MiniTableField_IsMap(field)) {
        const upb_Map* map = upb_Message_GetMap(src, field);
        if (map != NULL) {
          if (!upb_Message_Map_DeepClone(map, mini_table, field, dst, arena)) {
            return NULL;
          }
        }
      } else {
        const upb_Array* array = upb_Message_GetArray(src, field);
        if (array != NULL) {
          if (!upb_Message_Array_DeepClone(array, mini_table, field, dst,
                                           arena)) {
            return NULL;
          }
        }
      }
    }
  }
  // Clone extensions.
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(src);
  if (!in) return dst;

  for (size_t i = 0; i < in->size; i++) {
    upb_TaggedAuxPtr tagged_ptr = in->aux_data[i];
    if (upb_TaggedAuxPtr_IsCanonicalExtension(tagged_ptr)) {
      // Clone extension
      const upb_Extension* msg_ext =
          upb_TaggedAuxPtr_CanonicalExtension(tagged_ptr);
      const upb_MiniTableField* field = &msg_ext->ext->UPB_PRIVATE(field);
      upb_Extension* dst_ext = UPB_PRIVATE(_upb_Message_GetOrCreateExtension)(
          dst, msg_ext->ext, arena);
      if (!dst_ext) return NULL;
      if (upb_MiniTableField_IsScalar(field)) {
        if (!upb_Clone_ExtensionValue(msg_ext->ext, msg_ext, dst_ext, arena)) {
          return NULL;
        }
      } else {
        upb_Array* msg_array = (upb_Array*)msg_ext->data.array_val;
        UPB_ASSERT(msg_array);
        upb_Array* cloned_array = upb_Array_DeepClone(
            msg_array, upb_MiniTableField_CType(field),
            upb_MiniTableExtension_GetSubMessage(msg_ext->ext), arena);
        if (!cloned_array) {
          return NULL;
        }
        dst_ext->data.array_val = cloned_array;
      }
    } else if (upb_TaggedAuxPtr_IsUnknownStringView(tagged_ptr)) {
      // Clone unknown
      upb_StringView* unknown = upb_TaggedPtrAux_StringViewRepr(tagged_ptr);
      // Make a copy into destination arena.
      if (!UPB_PRIVATE(_upb_Message_AddUnknown)(
              dst, unknown->data, unknown->size, arena, kUpb_AddUnknown_Copy)) {
        return NULL;
      }
    }
  }

  return dst;
}

bool upb_Message_DeepCopy(upb_Message* dst, const upb_Message* src,
                          const upb_MiniTable* mini_table, upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(dst));
  upb_Message_Clear(dst, mini_table);
  return _upb_Message_Copy(dst, src, mini_table, arena) != NULL;
}

// Deep clones a message using the provided target arena.
//
// Returns NULL on failure.
upb_Message* upb_Message_DeepClone(const upb_Message* msg,
                                   const upb_MiniTable* m, upb_Arena* arena) {
  upb_Message* clone = upb_Message_New(m, arena);
  return _upb_Message_Copy(clone, msg, m, arena);
}

// Performs a shallow copy.
bool upb_Message_ShallowCopy(upb_Message* dst, const upb_Message* src,
                             const upb_MiniTable* m, upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(dst));
  memcpy(dst, src, m->UPB_PRIVATE(size));

  const upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(src);
  if (!in) return true;

  size_t size = UPB_SIZEOF_FLEX(upb_Message_Internal, aux_data, in->size);
  upb_Message_Internal* dst_in = upb_Arena_Malloc(arena, size);
  if (!dst_in) return false;

  dst_in->size = in->size;
  dst_in->capacity = in->size;

  for (size_t i = 0; i < in->size; i++) {
    upb_TaggedAux aux;
    switch (upb_TaggedAux_Get(in->aux_data[i], &aux)) {
      case kUpb_TaggedAuxType_CanonicalExtension: {
        const upb_Extension* msg_ext = aux.extension;
        upb_Extension* dst_ext = upb_Arena_Malloc(arena, sizeof(upb_Extension));
        if (!dst_ext) return false;
        *dst_ext = *msg_ext;
        dst_in->aux_data[i] = upb_TaggedAuxPtr_MakeCanonicalExtension(dst_ext);
        break;
      }
      case kUpb_TaggedAuxType_Unknown:
      case kUpb_TaggedAuxType_AliasedUnknown: {
        upb_StringView* dst_sv =
            upb_Arena_Malloc(arena, sizeof(upb_StringView));
        if (!dst_sv) return false;
        *dst_sv = *aux.unknown_data;
        dst_in->aux_data[i] = upb_TaggedAuxPtr_MakeUnknownDataAliased(dst_sv);
        break;
      }
      case kUpb_TaggedAuxType_NonCanonicalExtension: {
        const upb_Extension* msg_ext = aux.extension;
        upb_Extension* dst_ext = upb_Arena_Malloc(arena, sizeof(upb_Extension));
        if (!dst_ext) return false;
        *dst_ext = *msg_ext;
        dst_in->aux_data[i] =
            upb_TaggedAuxPtr_MakeNonCanonicalExtension(dst_ext);
        break;
      }
    }
  }

  UPB_PRIVATE(_upb_Message_SetInternal)(dst, dst_in);
  return true;
}

// Performs a shallow clone.
upb_Message* upb_Message_ShallowClone(const upb_Message* msg,
                                      const upb_MiniTable* m,
                                      upb_Arena* arena) {
  upb_Message* clone = upb_Message_New(m, arena);
  if (!clone) return NULL;
  if (!upb_Message_ShallowCopy(clone, msg, m, arena)) return NULL;
  return clone;
}


#include <stdint.h>
#include <stdlib.h>
#include <string.h>


// Must be last.

static int _upb_mapsorter_intkeys(const void* _a, const void* _b) {
  const upb_tabent* const* a = _a;
  const upb_tabent* const* b = _b;
  uintptr_t a_key = (*a)->key.num;
  uintptr_t b_key = (*b)->key.num;
  return a_key < b_key ? -1 : a_key > b_key;
}

static void _upb_mapsorter_getkeys(const void* _a, const void* _b, void* a_key,
                                   void* b_key, size_t size) {
  const upb_tabent* const* a = _a;
  const upb_tabent* const* b = _b;
  upb_StringView a_tabkey = upb_key_strview((*a)->key);
  upb_StringView b_tabkey = upb_key_strview((*b)->key);
  _upb_map_fromkey(a_tabkey, a_key, size);
  _upb_map_fromkey(b_tabkey, b_key, size);
}

static int _upb_mapsorter_cmpi64(const void* _a, const void* _b) {
  int64_t a, b;
  _upb_mapsorter_getkeys(_a, _b, &a, &b, 8);
  return a < b ? -1 : a > b;
}

static int _upb_mapsorter_cmpu64(const void* _a, const void* _b) {
  uint64_t a, b;
  _upb_mapsorter_getkeys(_a, _b, &a, &b, 8);
  return a < b ? -1 : a > b;
}

static int _upb_mapsorter_cmpi32(const void* _a, const void* _b) {
  int32_t a, b;
  _upb_mapsorter_getkeys(_a, _b, &a, &b, 4);
  return a < b ? -1 : a > b;
}

static int _upb_mapsorter_cmpu32(const void* _a, const void* _b) {
  uint32_t a, b;
  _upb_mapsorter_getkeys(_a, _b, &a, &b, 4);
  return a < b ? -1 : a > b;
}

static int _upb_mapsorter_cmpbool(const void* _a, const void* _b) {
  bool a, b;
  _upb_mapsorter_getkeys(_a, _b, &a, &b, 1);
  return a < b ? -1 : a > b;
}

static int _upb_mapsorter_cmpstr(const void* _a, const void* _b) {
  upb_StringView a, b;
  _upb_mapsorter_getkeys(_a, _b, &a, &b, UPB_MAPTYPE_STRING);
  size_t common_size = UPB_MIN(a.size, b.size);
  int cmp = memcmp(a.data, b.data, common_size);
  if (cmp) return -cmp;
  return a.size < b.size ? -1 : a.size > b.size;
}

static int (*const compar[kUpb_FieldType_SizeOf])(const void*, const void*) = {
    [kUpb_FieldType_Int64] = _upb_mapsorter_cmpi64,
    [kUpb_FieldType_SFixed64] = _upb_mapsorter_cmpi64,
    [kUpb_FieldType_SInt64] = _upb_mapsorter_cmpi64,

    [kUpb_FieldType_UInt64] = _upb_mapsorter_cmpu64,
    [kUpb_FieldType_Fixed64] = _upb_mapsorter_cmpu64,

    [kUpb_FieldType_Int32] = _upb_mapsorter_cmpi32,
    [kUpb_FieldType_SInt32] = _upb_mapsorter_cmpi32,
    [kUpb_FieldType_SFixed32] = _upb_mapsorter_cmpi32,
    [kUpb_FieldType_Enum] = _upb_mapsorter_cmpi32,

    [kUpb_FieldType_UInt32] = _upb_mapsorter_cmpu32,
    [kUpb_FieldType_Fixed32] = _upb_mapsorter_cmpu32,

    [kUpb_FieldType_Bool] = _upb_mapsorter_cmpbool,

    [kUpb_FieldType_String] = _upb_mapsorter_cmpstr,
    [kUpb_FieldType_Bytes] = _upb_mapsorter_cmpstr,
};

static bool _upb_mapsorter_resize(_upb_mapsorter* s, _upb_sortedmap* sorted,
                                  int size) {
  sorted->start = s->size;
  sorted->pos = sorted->start;
  sorted->end = sorted->start + size;

  if (sorted->end > s->cap) {
    const int oldsize = s->cap * sizeof(*s->entries);
    s->cap = upb_RoundUpToPowerOfTwo(sorted->end);
    const int newsize = s->cap * sizeof(*s->entries);
    s->entries = upb_grealloc(s->entries, oldsize, newsize);
    if (!s->entries) return false;
  }

  s->size = sorted->end;
  return true;
}

bool _upb_mapsorter_pushmap(_upb_mapsorter* s, upb_FieldType key_type,
                            const upb_Map* map, _upb_sortedmap* sorted) {
  int map_size = _upb_Map_Size(map);

  if (!_upb_mapsorter_resize(s, sorted, map_size)) return false;

  // Copy non-empty entries from the table to s->entries.
  const void** dst = &s->entries[sorted->start];
  const upb_tabent* src;
  const upb_tabent* end;
  if (map->UPB_PRIVATE(is_strtable)) {
    src = map->t.strtable.t.entries;
    end = src + upb_table_size(&map->t.strtable.t);
  } else {
    src = map->t.inttable.t.entries;
    end = src + upb_table_size(&map->t.inttable.t);
  }
  for (; src < end; src++) {
    if (!upb_tabent_isempty(src)) {
      *dst = src;
      dst++;
    }
  }
  UPB_ASSERT(dst == &s->entries[sorted->end]);

  // Sort entries according to the key type.
  qsort(&s->entries[sorted->start], map_size, sizeof(*s->entries),
        map->UPB_PRIVATE(is_strtable) ? compar[key_type]
                                      : _upb_mapsorter_intkeys);
  return true;
}

static int _upb_mapsorter_cmpext(const void* _a, const void* _b) {
  const upb_Extension* const* a = _a;
  const upb_Extension* const* b = _b;
  uint32_t a_num = upb_MiniTableExtension_Number((*a)->ext);
  uint32_t b_num = upb_MiniTableExtension_Number((*b)->ext);
  UPB_ASSERT(a_num != b_num);
  return a_num < b_num ? -1 : 1;
}

bool _upb_mapsorter_pushexts(_upb_mapsorter* s, const upb_Message_Internal* in,
                             _upb_sortedmap* sorted) {
  size_t count = 0;
  for (size_t i = 0; i < in->size; i++) {
    count += upb_TaggedAuxPtr_IsCanonicalExtension(in->aux_data[i]);
  }
  if (!_upb_mapsorter_resize(s, sorted, count)) return false;
  if (count == 0) return true;
  const upb_Extension** entry =
      (const upb_Extension**)&s->entries[sorted->start];
  for (size_t i = 0; i < in->size; i++) {
    upb_TaggedAuxPtr tagged_ptr = in->aux_data[i];
    if (upb_TaggedAuxPtr_IsCanonicalExtension(tagged_ptr)) {
      *entry++ = upb_TaggedAuxPtr_CanonicalExtension(tagged_ptr);
    }
  }
  qsort(&s->entries[sorted->start], count, sizeof(*s->entries),
        _upb_mapsorter_cmpext);
  return true;
}


#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Must be last.

static upb_FindUnknownRet2 upb_FindUnknownRet2_ParseError(void) {
  return (upb_FindUnknownRet2){.status = kUpb_FindUnknown_ParseError};
}

upb_FindUnknownRet2 upb_Message_FindUnknown2(const struct upb_Message* msg,
                                             uint32_t field_number,
                                             int depth_limit) {
  depth_limit = depth_limit ? depth_limit : 100;
  upb_FindUnknownRet2 ret;
  ret.iter = kUpb_Message_UnknownBegin;
  while (upb_Message_NextUnknown2(msg, &ret.unknown, &ret.iter)) {
    if (ret.unknown.type == kUpb_MessageUnknownType_StringView) {
      upb_EpsCopyInputStream stream;
      const char* ptr = ret.unknown.value.bytes.data;
      upb_EpsCopyInputStream_Init(&stream, &ptr, ret.unknown.value.bytes.size);

      while (!upb_EpsCopyInputStream_IsDone(&stream, &ptr)) {
        uint32_t tag;
        const char* unknown_begin = ptr;
        ptr = upb_WireReader_ReadTag(ptr, &tag, &stream);
        if (!ptr) return upb_FindUnknownRet2_ParseError();
        if (field_number == upb_WireReader_GetFieldNumber(tag)) {
          upb_StringView data;
          ret.status = kUpb_FindUnknown_Ok;
          upb_EpsCopyCapture capture;
          upb_EpsCopyCapture_Start(&capture, &stream, unknown_begin);
          ptr = _upb_WireReader_SkipValue(ptr, tag, depth_limit, &stream);
          if (!ptr || !upb_EpsCopyCapture_End(&capture, &stream, ptr, &data)) {
            return upb_FindUnknownRet2_ParseError();
          }
          ret.unknown.value.bytes = data;
          return ret;
        }

        ptr = _upb_WireReader_SkipValue(ptr, tag, depth_limit, &stream);
        if (!ptr) return upb_FindUnknownRet2_ParseError();
      }
    } else if (ret.unknown.type ==
               kUpb_MessageUnknownType_NonCanonicalExtension) {
      uint32_t ext_field_number =
          upb_MiniTableExtension_Number(ret.unknown.value.extension->ext);
      if (ext_field_number == field_number) {
        ret.status = kUpb_FindUnknown_Ok;
        return ret;
      }
    }
  }
  ret.status = kUpb_FindUnknown_NotPresent;
  ret.unknown.type = kUpb_MessageUnknownType_StringView;
  ret.unknown.value.bytes.data = NULL;
  ret.unknown.value.bytes.size = 0;
  ret.iter = kUpb_Message_UnknownBegin;
  return ret;
}

upb_Message_DeleteUnknownStatus upb_Message_DeleteUnknown2(
    struct upb_Message* msg, struct upb_MessageUnknown* data, uintptr_t* iter,
    struct upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  UPB_ASSERT(*iter != kUpb_Message_UnknownBegin);
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  UPB_ASSERT(in);
  UPB_ASSERT(*iter <= in->size);
  upb_TaggedAuxPtr unknown_ptr = in->aux_data[*iter - 1];

  if (data->type == kUpb_MessageUnknownType_NonCanonicalExtension) {
    UPB_ASSERT(upb_TaggedAuxPtr_IsNonCanonicalExtension(unknown_ptr));
    // When the unknown is a non-canonical extension, we just remove it from the
    // aux data array.
    in->aux_data[*iter - 1] = upb_TaggedAuxPtr_Null();
    return upb_Message_NextUnknown2(msg, data, iter)
               ? kUpb_DeleteUnknown_IterUpdated
               : kUpb_DeleteUnknown_DeletedLast;
  }

  UPB_ASSERT(upb_TaggedAuxPtr_IsUnknownStringView(unknown_ptr));
  upb_StringView* unknown = upb_TaggedPtrAux_StringViewRepr(unknown_ptr);
  UPB_ASSERT(data->type == kUpb_MessageUnknownType_StringView);
  upb_StringView* data_bytes = &data->value.bytes;
  if (unknown->data == data_bytes->data && unknown->size == data_bytes->size) {
    // Remove whole field
    in->aux_data[*iter - 1] = upb_TaggedAuxPtr_Null();
  } else if (unknown->data == data_bytes->data) {
    // Strip prefix
    unknown->data += data_bytes->size;
    unknown->size -= data_bytes->size;
    *data_bytes = *unknown;
    return kUpb_DeleteUnknown_IterUpdated;
  } else if (unknown->data + unknown->size ==
             data_bytes->data + data_bytes->size) {
    // Truncate existing field
    unknown->size -= data_bytes->size;
    if (!upb_TaggedAuxPtr_IsUnknownAliased(unknown_ptr)) {
      in->aux_data[*iter - 1] =
          upb_TaggedAuxPtr_MakeUnknownDataAliased(unknown);
    }
  } else {
    UPB_ASSERT(unknown->data < data_bytes->data &&
               unknown->data + unknown->size >
                   data_bytes->data + data_bytes->size);
    // Split in the middle
    upb_StringView* prefix = unknown;
    upb_StringView* suffix = upb_Arena_Malloc(arena, sizeof(upb_StringView));
    if (!suffix) {
      return kUpb_DeleteUnknown_AllocFail;
    }
    if (!UPB_PRIVATE(_upb_Message_ReserveSlot)(msg, arena)) {
      return kUpb_DeleteUnknown_AllocFail;
    }
    in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
    if (*iter != in->size) {
      // Shift later entries down so that unknown field ordering is preserved
      memmove(&in->aux_data[*iter + 1], &in->aux_data[*iter],
              sizeof(upb_TaggedAuxPtr) * (in->size - *iter));
    }
    in->aux_data[*iter] = upb_TaggedAuxPtr_MakeUnknownDataAliased(suffix);
    if (!upb_TaggedAuxPtr_IsUnknownAliased(unknown_ptr)) {
      in->aux_data[*iter - 1] = upb_TaggedAuxPtr_MakeUnknownDataAliased(prefix);
    }
    in->size++;
    suffix->data = data_bytes->data + data_bytes->size;
    suffix->size = (prefix->data + prefix->size) - suffix->data;
    prefix->size = data_bytes->data - prefix->data;
  }
  return upb_Message_NextUnknown2(msg, data, iter)
             ? kUpb_DeleteUnknown_IterUpdated
             : kUpb_DeleteUnknown_DeletedLast;
}


#include <stdint.h>
#include <string.h>


// Must be last.

// Strings/bytes are special-cased in maps.
char _upb_Map_CTypeSizeTable[12] = {
    [kUpb_CType_Bool] = 1,
    [kUpb_CType_Float] = 4,
    [kUpb_CType_Int32] = 4,
    [kUpb_CType_UInt32] = 4,
    [kUpb_CType_Enum] = 4,
    [kUpb_CType_Message] = sizeof(void*),
    [kUpb_CType_Double] = 8,
    [kUpb_CType_Int64] = 8,
    [kUpb_CType_UInt64] = 8,
    [kUpb_CType_String] = UPB_MAPTYPE_STRING,
    [kUpb_CType_Bytes] = UPB_MAPTYPE_STRING,
};

upb_Map* upb_Map_New(upb_Arena* a, upb_CType key_type, upb_CType value_type) {
  return _upb_Map_New(a, _upb_Map_CTypeSize(key_type),
                      _upb_Map_CTypeSize(value_type));
}

size_t upb_Map_Size(const upb_Map* map) { return _upb_Map_Size(map); }

bool upb_Map_Get(const upb_Map* map, upb_MessageValue key,
                 upb_MessageValue* val) {
  return _upb_Map_Get(map, &key, map->key_size, val, map->val_size);
}

struct upb_Message* upb_Map_GetMutable(upb_Map* map, upb_MessageValue key) {
  UPB_ASSERT(map->val_size == sizeof(upb_Message*));
  upb_Message* val = NULL;
  if (_upb_Map_Get(map, &key, map->key_size, &val, sizeof(upb_Message*))) {
    return val;
  } else {
    return NULL;
  }
}

void upb_Map_Clear(upb_Map* map) { _upb_Map_Clear(map); }

upb_MapInsertStatus upb_Map_Insert(upb_Map* map, upb_MessageValue key,
                                   upb_MessageValue val, upb_Arena* arena) {
  UPB_ASSERT(arena);
  return (upb_MapInsertStatus)_upb_Map_Insert(map, &key, map->key_size, &val,
                                              map->val_size, arena);
}

bool upb_Map_Delete(upb_Map* map, upb_MessageValue key, upb_MessageValue* val) {
  upb_value v;
  const bool removed = _upb_Map_Delete(map, &key, map->key_size, &v);
  if (val) _upb_map_fromvalue(v, val, map->val_size);
  return removed;
}

bool upb_Map_Next(const upb_Map* map, upb_MessageValue* key,
                  upb_MessageValue* val, size_t* iter) {
  upb_value v;
  bool ret;
  if (map->UPB_PRIVATE(is_strtable)) {
    upb_StringView strkey;
    ret = upb_strtable_next2(&map->t.strtable, &strkey, &v, (intptr_t*)iter);
    if (ret) {
      _upb_map_fromkey(strkey, key, map->key_size);
    }
  } else {
    uintptr_t intkey;
    ret = upb_inttable_next(&map->t.inttable, &intkey, &v, (intptr_t*)iter);
    if (ret) {
      memcpy(key, &intkey, map->key_size);
    }
  }
  if (ret) {
    _upb_map_fromvalue(v, val, map->val_size);
  }
  return ret;
}

UPB_API void upb_Map_SetEntryValue(upb_Map* map, size_t iter,
                                   upb_MessageValue val) {
  upb_value v;
  _upb_map_tovalue(&val, map->val_size, &v, NULL);
  if (map->UPB_PRIVATE(is_strtable)) {
    upb_strtable_setentryvalue(&map->t.strtable, iter, v);
  } else {
    upb_inttable_setentryvalue(&map->t.inttable, iter, v);
  }
}

bool upb_MapIterator_Next(const upb_Map* map, size_t* iter) {
  return _upb_map_next(map, iter);
}

bool upb_MapIterator_Done(const upb_Map* map, size_t iter) {
  UPB_ASSERT(iter != kUpb_Map_Begin);
  if (map->UPB_PRIVATE(is_strtable)) {
    upb_strtable_iter i;
    i.t = &map->t.strtable;
    i.index = iter;
    return upb_strtable_done(&i);
  } else {
    return upb_inttable_done(&map->t.inttable, iter);
  }
}

// Returns the key and value for this entry of the map.
upb_MessageValue upb_MapIterator_Key(const upb_Map* map, size_t iter) {
  upb_MessageValue ret;
  if (map->UPB_PRIVATE(is_strtable)) {
    upb_strtable_iter i;
    i.t = &map->t.strtable;
    i.index = iter;
    _upb_map_fromkey(upb_strtable_iter_key(&i), &ret, map->key_size);
  } else {
    uintptr_t intkey = upb_inttable_iter_key(&map->t.inttable, iter);
    memcpy(&ret, &intkey, map->key_size);
  }
  return ret;
}

upb_MessageValue upb_MapIterator_Value(const upb_Map* map, size_t iter) {
  upb_value v;
  if (map->UPB_PRIVATE(is_strtable)) {
    upb_strtable_iter i;
    i.t = &map->t.strtable;
    i.index = iter;
    v = upb_strtable_iter_value(&i);
  } else {
    v = upb_inttable_iter_value(&map->t.inttable, iter);
  }

  upb_MessageValue ret;
  _upb_map_fromvalue(v, &ret, map->val_size);
  return ret;
}

void upb_Map_Freeze(upb_Map* map, const upb_MiniTable* m) {
  if (upb_Map_IsFrozen(map)) return;
  UPB_PRIVATE(_upb_Map_ShallowFreeze)(map);

  if (m) {
    size_t iter = kUpb_Map_Begin;
    upb_MessageValue key, val;

    while (upb_Map_Next(map, &key, &val, &iter)) {
      upb_Message_Freeze((upb_Message*)val.msg_val, m);
    }
  }
}

// EVERYTHING BELOW THIS LINE IS INTERNAL - DO NOT USE /////////////////////////

upb_Map* _upb_Map_New(upb_Arena* a, size_t key_size, size_t value_size) {
  upb_Map* map = upb_Arena_Malloc(a, sizeof(upb_Map));
  if (!map) return NULL;

  if (key_size <= sizeof(uintptr_t) && key_size != UPB_MAPTYPE_STRING) {
    if (!upb_inttable_init(&map->t.inttable, a)) return NULL;
    map->UPB_PRIVATE(is_strtable) = false;
  } else {
    if (!upb_strtable_init(&map->t.strtable, 4, a)) return NULL;
    map->UPB_PRIVATE(is_strtable) = true;
  }
  map->key_size = key_size;
  map->val_size = value_size;
  map->UPB_PRIVATE(is_frozen) = false;

  return map;
}



// Must be last.

bool upb_Message_SetMapEntry(upb_Map* map, const upb_MiniTableField* f,
                             upb_Message* map_entry_message, upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(map_entry_message));
  const upb_MiniTable* map_entry_mini_table =
      upb_MiniTable_MapEntrySubMessage(f);
  UPB_ASSERT(map_entry_mini_table);
  const upb_MiniTableField* map_entry_key_field =
      upb_MiniTable_MapKey(map_entry_mini_table);
  const upb_MiniTableField* map_entry_value_field =
      upb_MiniTable_MapValue(map_entry_mini_table);
  // Map key/value cannot have explicit defaults,
  // hence assuming a zero default is valid.
  upb_MessageValue default_val = upb_MessageValue_Zero();
  upb_MessageValue map_entry_key =
      upb_Message_GetField(map_entry_message, map_entry_key_field, default_val);
  upb_MessageValue map_entry_value = upb_Message_GetField(
      map_entry_message, map_entry_value_field, default_val);
  return upb_Map_Set(map, map_entry_key, map_entry_value, arena);
}


#include <stdbool.h>
#include <stdint.h>
#include <string.h>


// Must be last.

// Parses unknown data by merging into existing base_message or creating a
// new message using mini_table.
static upb_UnknownToMessageRet upb_MiniTable_ParseUnknownMessage(
    const char* unknown_data, size_t unknown_size,
    const upb_MiniTable* mini_table, upb_Message* base_message,
    int decode_options, upb_Arena* arena) {
  upb_UnknownToMessageRet ret;
  ret.message =
      base_message ? base_message : _upb_Message_New(mini_table, arena);
  if (!ret.message) {
    ret.status = kUpb_UnknownToMessage_OutOfMemory;
    return ret;
  }
  // Decode sub message using unknown field contents.
  const char* data = unknown_data;
  uint32_t tag;
  uint64_t message_len = 0;
  data = upb_WireReader_ReadTag(data, &tag, NULL);
  data = upb_WireReader_ReadVarint(data, &message_len, NULL);
  upb_DecodeStatus status = upb_Decode(data, message_len, ret.message,
                                       mini_table, NULL, decode_options, arena);
  if (status == kUpb_DecodeStatus_OutOfMemory) {
    ret.status = kUpb_UnknownToMessage_OutOfMemory;
  } else if (status == kUpb_DecodeStatus_Ok) {
    ret.status = kUpb_UnknownToMessage_Ok;
  } else {
    ret.status = kUpb_UnknownToMessage_ParseError;
  }
  return ret;
}

upb_GetExtension_Status upb_Message_GetOrPromoteExtension(
    upb_Message* msg, const upb_MiniTableExtension* ext_table,
    int decode_options, upb_Arena* arena, upb_MessageValue* value) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  UPB_ASSERT(upb_MiniTableExtension_CType(ext_table) == kUpb_CType_Message);
  const upb_Extension* extension =
      UPB_PRIVATE(_upb_Message_Getext)(msg, ext_table);
  if (extension) {
    memcpy(value, &extension->data, sizeof(upb_MessageValue));
    return kUpb_GetExtension_Ok;
  }

  // Check unknown fields, if available promote.
  int found_count = 0;
  uint32_t field_number = upb_MiniTableExtension_Number(ext_table);
  const upb_MiniTable* extension_table =
      upb_MiniTableExtension_GetSubMessage(ext_table);
  // Will be populated on first parse and then reused
  upb_Message* extension_msg = NULL;
  int depth_limit = 100;
  uintptr_t iter = kUpb_Message_UnknownBegin;
  upb_StringView data;
  while (upb_Message_NextUnknown(msg, &data, &iter)) {
    const char* ptr = data.data;
    upb_EpsCopyInputStream stream;
    upb_EpsCopyInputStream_Init(&stream, &ptr, data.size);
    while (!upb_EpsCopyInputStream_IsDone(&stream, &ptr)) {
      uint32_t tag;
      const char* unknown_begin = ptr;
      ptr = upb_WireReader_ReadTag(ptr, &tag, &stream);
      if (!ptr) return kUpb_GetExtension_ParseError;
      if (field_number == upb_WireReader_GetFieldNumber(tag)) {
        upb_StringView data;
        found_count++;
        upb_EpsCopyCapture capture;
        upb_EpsCopyCapture_Start(&capture, &stream, unknown_begin);
        ptr = _upb_WireReader_SkipValue(ptr, tag, depth_limit, &stream);
        if (!ptr || !upb_EpsCopyCapture_End(&capture, &stream, ptr, &data)) {
          return kUpb_GetExtension_ParseError;
        }
        upb_UnknownToMessageRet parse_result =
            upb_MiniTable_ParseUnknownMessage(
                data.data, data.size, extension_table,
                /* base_message= */ extension_msg, decode_options, arena);
        switch (parse_result.status) {
          case kUpb_UnknownToMessage_OutOfMemory:
            return kUpb_GetExtension_OutOfMemory;
          case kUpb_UnknownToMessage_ParseError:
            return kUpb_GetExtension_ParseError;
          case kUpb_UnknownToMessage_NotFound:
            return kUpb_GetExtension_NotPresent;
          case kUpb_UnknownToMessage_Ok:
            extension_msg = parse_result.message;
        }
      } else {
        ptr = _upb_WireReader_SkipValue(ptr, tag, depth_limit, &stream);
        if (!ptr) return kUpb_GetExtension_ParseError;
      }
    }
  }
  if (!extension_msg) {
    return kUpb_GetExtension_NotPresent;
  }

  upb_Extension* ext =
      UPB_PRIVATE(_upb_Message_GetOrCreateExtension)(msg, ext_table, arena);
  if (!ext) {
    return kUpb_GetExtension_OutOfMemory;
  }
  ext->data.msg_val = extension_msg;

  while (found_count > 0) {
    upb_FindUnknownRet found = upb_Message_FindUnknown(msg, field_number, 0);
    UPB_ASSERT(found.status == kUpb_FindUnknown_Ok);
    upb_StringView view = {.data = found.ptr, .size = found.len};
    if (upb_Message_DeleteUnknown(msg, &view, &found.iter, arena) ==
        kUpb_DeleteUnknown_AllocFail) {
      return kUpb_GetExtension_OutOfMemory;
    }
    found_count--;
  }
  value->msg_val = extension_msg;
  return kUpb_GetExtension_Ok;
}

static upb_FindUnknownRet upb_FindUnknownRet_ParseError(void) {
  return (upb_FindUnknownRet){.status = kUpb_FindUnknown_ParseError};
}

upb_FindUnknownRet upb_Message_FindUnknown(const upb_Message* msg,
                                           uint32_t field_number,
                                           int depth_limit) {
  depth_limit = depth_limit ? depth_limit : 100;
  upb_FindUnknownRet ret;
  ret.iter = kUpb_Message_UnknownBegin;
  upb_StringView data;
  while (upb_Message_NextUnknown(msg, &data, &ret.iter)) {
    upb_EpsCopyInputStream stream;
    const char* ptr = data.data;
    upb_EpsCopyInputStream_Init(&stream, &ptr, data.size);

    while (!upb_EpsCopyInputStream_IsDone(&stream, &ptr)) {
      uint32_t tag;
      const char* unknown_begin = ptr;
      ptr = upb_WireReader_ReadTag(ptr, &tag, &stream);
      if (!ptr) return upb_FindUnknownRet_ParseError();
      if (field_number == upb_WireReader_GetFieldNumber(tag)) {
        upb_StringView data;
        ret.status = kUpb_FindUnknown_Ok;
        upb_EpsCopyCapture capture;
        upb_EpsCopyCapture_Start(&capture, &stream, unknown_begin);
        ptr = _upb_WireReader_SkipValue(ptr, tag, depth_limit, &stream);
        if (!ptr || !upb_EpsCopyCapture_End(&capture, &stream, ptr, &data)) {
          return upb_FindUnknownRet_ParseError();
        }
        ret.ptr = data.data;
        ret.len = data.size;
        return ret;
      }

      ptr = _upb_WireReader_SkipValue(ptr, tag, depth_limit, &stream);
      if (!ptr) return upb_FindUnknownRet_ParseError();
    }
  }
  ret.status = kUpb_FindUnknown_NotPresent;
  ret.ptr = NULL;
  ret.len = 0;
  ret.iter = kUpb_Message_UnknownBegin;
  return ret;
}

// Warning: See TODO
upb_UnknownToMessageRet upb_MiniTable_PromoteUnknownToMessage(
    upb_Message* msg, const upb_MiniTable* mini_table,
    const upb_MiniTableField* field, const upb_MiniTable* sub_mini_table,
    int decode_options, upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  upb_FindUnknownRet unknown;
  // We need to loop and merge unknowns that have matching tag field->number.
  upb_Message* message = NULL;
  // Callers should check that message is not set first before calling
  // PromotoUnknownToMessage.
  UPB_ASSERT(upb_MiniTable_GetSubMessageTable(field) == sub_mini_table);
  bool is_oneof = upb_MiniTableField_IsInOneof(field);
  if (!is_oneof || UPB_PRIVATE(_upb_Message_GetOneofCase)(msg, field) ==
                       upb_MiniTableField_Number(field)) {
    UPB_ASSERT(upb_Message_GetMessage(msg, field) == NULL);
  }
  upb_UnknownToMessageRet ret;
  ret.status = kUpb_UnknownToMessage_Ok;
  do {
    unknown = upb_Message_FindUnknown(
        msg, upb_MiniTableField_Number(field),
        upb_DecodeOptions_GetEffectiveMaxDepth(decode_options));
    switch (unknown.status) {
      case kUpb_FindUnknown_Ok: {
        const char* unknown_data = unknown.ptr;
        size_t unknown_size = unknown.len;
        ret = upb_MiniTable_ParseUnknownMessage(unknown_data, unknown_size,
                                                sub_mini_table, message,
                                                decode_options, arena);
        if (ret.status == kUpb_UnknownToMessage_Ok) {
          message = ret.message;
          upb_StringView del =
              upb_StringView_FromDataAndSize(unknown_data, unknown_size);
          upb_Message_DeleteUnknownStatus del_status =
              upb_Message_DeleteUnknown(msg, &del, &(unknown.iter), arena);
          if (del_status == kUpb_DeleteUnknown_AllocFail) {
            ret.status = kUpb_UnknownToMessage_OutOfMemory;
            return ret;
          }
        }
      } break;
      case kUpb_FindUnknown_ParseError:
        ret.status = kUpb_UnknownToMessage_ParseError;
        break;
      case kUpb_FindUnknown_NotPresent:
        // If we parsed at least one unknown, we are done.
        ret.status =
            message ? kUpb_UnknownToMessage_Ok : kUpb_UnknownToMessage_NotFound;
        break;
    }
  } while (unknown.status == kUpb_FindUnknown_Ok);
  if (message) {
    if (is_oneof) {
      UPB_PRIVATE(_upb_Message_SetOneofCase)(msg, field);
    }
    upb_Message_SetMessage(msg, field, message);
    ret.message = message;
  }
  return ret;
}

// Moves repeated messages in unknowns to a upb_Array.
//
// Since the repeated field is not a scalar type we don't check for
// kUpb_LabelFlags_IsPacked.
// TODO: Optimize. Instead of converting messages one at a time,
// scan all unknown data once and compact.
upb_UnknownToMessage_Status upb_MiniTable_PromoteUnknownToMessageArray(
    upb_Message* msg, const upb_MiniTableField* field,
    const upb_MiniTable* mini_table, int decode_options, upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));

  upb_Array* repeated_messages = upb_Message_GetMutableArray(msg, field);
  // Find all unknowns with given field number and parse.
  upb_FindUnknownRet unknown;
  do {
    unknown = upb_Message_FindUnknown(
        msg, upb_MiniTableField_Number(field),
        upb_DecodeOptions_GetEffectiveMaxDepth(decode_options));
    if (unknown.status == kUpb_FindUnknown_Ok) {
      upb_UnknownToMessageRet ret = upb_MiniTable_ParseUnknownMessage(
          unknown.ptr, unknown.len, mini_table,
          /* base_message= */ NULL, decode_options, arena);
      if (ret.status == kUpb_UnknownToMessage_Ok) {
        upb_MessageValue value;
        value.msg_val = ret.message;
        // Allocate array on demand before append.
        if (!repeated_messages) {
          if (!upb_Message_ResizeArrayUninitialized(msg, field, 0, arena)) {
            return kUpb_UnknownToMessage_OutOfMemory;
          }
          repeated_messages = upb_Message_GetMutableArray(msg, field);
        }
        if (!upb_Array_Append(repeated_messages, value, arena)) {
          return kUpb_UnknownToMessage_OutOfMemory;
        }
        upb_StringView del =
            upb_StringView_FromDataAndSize(unknown.ptr, unknown.len);
        upb_Message_DeleteUnknownStatus del_status =
            upb_Message_DeleteUnknown(msg, &del, &(unknown.iter), arena);
        if (del_status == kUpb_DeleteUnknown_AllocFail) {
          return kUpb_UnknownToMessage_OutOfMemory;
        }
      } else {
        return ret.status;
      }
    }
  } while (unknown.status == kUpb_FindUnknown_Ok);
  return kUpb_UnknownToMessage_Ok;
}

// Moves repeated messages in unknowns to a upb_Map.
upb_UnknownToMessage_Status upb_MiniTable_PromoteUnknownToMap(
    upb_Message* msg, const upb_MiniTable* mini_table,
    const upb_MiniTableField* field, int decode_options, upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));

  const upb_MiniTable* map_entry_mini_table =
      upb_MiniTable_MapEntrySubMessage(field);
  UPB_ASSERT(upb_MiniTable_FieldCount(map_entry_mini_table) == 2);
  // Find all unknowns with given field number and parse.
  upb_FindUnknownRet unknown;
  while (1) {
    unknown = upb_Message_FindUnknown(
        msg, upb_MiniTableField_Number(field),
        upb_DecodeOptions_GetEffectiveMaxDepth(decode_options));
    if (unknown.status != kUpb_FindUnknown_Ok) break;
    upb_UnknownToMessageRet ret = upb_MiniTable_ParseUnknownMessage(
        unknown.ptr, unknown.len, map_entry_mini_table,
        /* base_message= */ NULL, decode_options, arena);
    if (ret.status != kUpb_UnknownToMessage_Ok) return ret.status;
    // Allocate map on demand before append.
    upb_Map* map = upb_Message_GetOrCreateMutableMap(msg, map_entry_mini_table,
                                                     field, arena);
    upb_Message* map_entry_message = ret.message;
    bool insert_success =
        upb_Message_SetMapEntry(map, field, map_entry_message, arena);
    if (!insert_success) return kUpb_UnknownToMessage_OutOfMemory;
    upb_StringView del =
        upb_StringView_FromDataAndSize(unknown.ptr, unknown.len);
    upb_Message_DeleteUnknownStatus del_status =
        upb_Message_DeleteUnknown(msg, &del, &unknown.iter, arena);
    if (del_status == kUpb_DeleteUnknown_AllocFail) {
      return kUpb_UnknownToMessage_OutOfMemory;
    }
  }
  return kUpb_UnknownToMessage_Ok;
}

const char* upb_FindUnknownStatus_String(upb_FindUnknown_Status status) {
  switch (status) {
    case kUpb_FindUnknown_Ok:
      return "Ok";
    case kUpb_FindUnknown_ParseError:
      return "Parse error";
    case kUpb_FindUnknown_NotPresent:
      return "Field not found";
    default:
      return "Unknown status";
  }
}


#include <stdarg.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Must be last.

upb_Message* upb_Message_New(const upb_MiniTable* m, upb_Arena* a) {
  return _upb_Message_New(m, a);
}

UPB_NOINLINE bool UPB_PRIVATE(_upb_Message_AddUnknownSlowPath)(upb_Message* msg,
                                                               const char* data,
                                                               size_t len,
                                                               upb_Arena* arena,
                                                               bool alias) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  {
    upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
    // Alias fast path was already checked in the inline function that calls
    // this one
    if (!alias && in && in->size) {
      upb_TaggedAuxPtr ptr = in->aux_data[in->size - 1];
      if (upb_TaggedAuxPtr_IsUnknownStringView(ptr)) {
        upb_StringView* existing = upb_TaggedPtrAux_StringViewRepr(ptr);
        if (!upb_TaggedAuxPtr_IsUnknownAliased(ptr)) {
          // If part of the existing field was deleted at the beginning, we can
          // reconstruct it by comparing the address of the end with the address
          // of the entry itself; having the non-aliased tag means that the
          // string_view and the data it points to are part of the same original
          // upb_Arena_Malloc allocation, and the end of the string view
          // represents the end of that allocation.
          size_t prev_alloc_size =
              (existing->data + existing->size) - (char*)existing;
          if (SIZE_MAX - prev_alloc_size >= len) {
            size_t new_alloc_size = prev_alloc_size + len;
            if (upb_Arena_TryExtend(arena, existing, prev_alloc_size,
                                    new_alloc_size)) {
              memcpy(UPB_PTR_AT(existing, prev_alloc_size, void), data, len);
              existing->size += len;
              return true;
            }
          }
        }
      }
    }
  }
  // TODO: b/376969853  - Add debug check that the unknown field is an overall
  // valid proto field
  if (!UPB_PRIVATE(_upb_Message_ReserveSlot)(msg, arena)) {
    return false;
  }
  upb_StringView* view;
  if (alias) {
    view = upb_Arena_Malloc(arena, sizeof(upb_StringView));
    if (!view) return false;
    view->data = data;
  } else {
    if (SIZE_MAX - sizeof(upb_StringView) < len) return false;
    view = upb_Arena_Malloc(arena, sizeof(upb_StringView) + len);
    if (!view) return false;
    char* copy = UPB_PTR_AT(view, sizeof(upb_StringView), char);
    memcpy(copy, data, len);
    view->data = copy;
  }
  view->size = len;
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  in->aux_data[in->size++] = alias
                                 ? upb_TaggedAuxPtr_MakeUnknownDataAliased(view)
                                 : upb_TaggedAuxPtr_MakeUnknownData(view);
  return true;
}

bool UPB_PRIVATE(_upb_Message_AddUnknownV)(struct upb_Message* msg,
                                           upb_Arena* arena,
                                           upb_StringView data[],
                                           size_t count) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  UPB_ASSERT(count > 0);
  size_t total_len = 0;
  for (size_t i = 0; i < count; i++) {
    if (SIZE_MAX - total_len < data[i].size) {
      return false;
    }
    total_len += data[i].size;
  }

  {
    upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
    if (in && in->size) {
      upb_TaggedAuxPtr ptr = in->aux_data[in->size - 1];
      if (upb_TaggedAuxPtr_IsUnknownStringView(ptr)) {
        upb_StringView* existing = upb_TaggedPtrAux_StringViewRepr(ptr);
        if (!upb_TaggedAuxPtr_IsUnknownAliased(ptr)) {
          size_t prev_alloc_size =
              (existing->data + existing->size) - (char*)existing;
          if (SIZE_MAX - prev_alloc_size >= total_len) {
            size_t new_alloc_size = prev_alloc_size + total_len;
            if (upb_Arena_TryExtend(arena, existing, prev_alloc_size,
                                    new_alloc_size)) {
              char* copy = UPB_PTR_AT(existing, prev_alloc_size, char);
              for (size_t i = 0; i < count; i++) {
                memcpy(copy, data[i].data, data[i].size);
                copy += data[i].size;
              }
              existing->size += total_len;
              return true;
            }
          }
        }
      }
    }
  }

  if (SIZE_MAX - sizeof(upb_StringView) < total_len) return false;
  if (!UPB_PRIVATE(_upb_Message_ReserveSlot)(msg, arena)) return false;

  upb_StringView* view =
      upb_Arena_Malloc(arena, sizeof(upb_StringView) + total_len);
  if (!view) return false;
  char* copy = UPB_PTR_AT(view, sizeof(upb_StringView), char);
  view->data = copy;
  view->size = total_len;
  for (size_t i = 0; i < count; i++) {
    memcpy(copy, data[i].data, data[i].size);
    copy += data[i].size;
  }
  // TODO: b/376969853  - Add debug check that the unknown field is an overall
  // valid proto field
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  in->aux_data[in->size++] = upb_TaggedAuxPtr_MakeUnknownData(view);
  return true;
}

void _upb_Message_DiscardUnknown_shallow(upb_Message* msg) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  if (!in) return;
  uint32_t size = 0;
  for (uint32_t i = 0; i < in->size; i++) {
    upb_TaggedAuxPtr tagged_ptr = in->aux_data[i];
    if (upb_TaggedAuxPtr_IsSemanticallyKnown(tagged_ptr)) {
      in->aux_data[size++] = tagged_ptr;
    }
  }
  in->size = size;
}

upb_Message_DeleteUnknownStatus upb_Message_DeleteUnknown(upb_Message* msg,
                                                          upb_StringView* data,
                                                          uintptr_t* iter,
                                                          upb_Arena* arena) {
  upb_MessageUnknown unknown;
  unknown.type = kUpb_MessageUnknownType_StringView;
  unknown.value.bytes = *data;

  upb_Message_DeleteUnknownStatus res =
      upb_Message_DeleteUnknown2(msg, &unknown, iter, arena);
  UPB_ASSERT(unknown.type == kUpb_MessageUnknownType_StringView);
  if (res == kUpb_DeleteUnknown_IterUpdated ||
      res == kUpb_DeleteUnknown_DeletedLast) {
    // the unknown data remains the same on the result of
    // kUpb_DeleteUnknown_AllocFail.
    *data = unknown.value.bytes;
  }
  return res;
}

size_t upb_Message_ExtensionCount(const upb_Message* msg) {
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  if (!in) return 0;
  const upb_MiniTableExtension* ext;
  upb_MessageValue val;
  uintptr_t iter = kUpb_Message_ExtensionBegin;
  size_t count = 0;
  while (upb_Message_NextExtension(msg, &ext, &val, &iter)) {
    count++;
  }
  return count;
}

void upb_Message_Freeze(upb_Message* msg, const upb_MiniTable* m) {
  if (upb_Message_IsFrozen(msg)) return;
  UPB_PRIVATE(_upb_Message_ShallowFreeze)(msg);

  // Base Fields.
  const size_t field_count = upb_MiniTable_FieldCount(m);

  for (size_t i = 0; i < field_count; i++) {
    const upb_MiniTableField* f = upb_MiniTable_GetFieldByIndex(m, i);
    const upb_MiniTable* m2 = upb_MiniTable_SubMessage(f);

    switch (UPB_PRIVATE(_upb_MiniTableField_Mode)(f)) {
      case kUpb_FieldMode_Array: {
        upb_Array* arr = upb_Message_GetMutableArray(msg, f);
        if (arr) upb_Array_Freeze(arr, m2);
        break;
      }
      case kUpb_FieldMode_Map: {
        upb_Map* map = upb_Message_GetMutableMap(msg, f);
        if (map) {
          const upb_MiniTableField* f2 = upb_MiniTable_MapValue(m2);
          const upb_MiniTable* m3 = upb_MiniTable_SubMessage(f2);
          upb_Map_Freeze(map, m3);
        }
        break;
      }
      case kUpb_FieldMode_Scalar: {
        if (m2) {
          upb_Message* msg2 = upb_Message_GetMutableMessage(msg, f);
          if (msg2) upb_Message_Freeze(msg2, m2);
        }
        break;
      }
    }
  }

  // Extensions.
  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  // TODO: b/376969853 - use iterator API
  uint32_t size = in ? in->size : 0;
  for (size_t i = 0; i < size; i++) {
    upb_TaggedAuxPtr tagged_ptr = in->aux_data[i];
    upb_TaggedAux aux;
    upb_TaggedAuxType type = upb_TaggedAux_Get(tagged_ptr, &aux);
    if (type != kUpb_TaggedAuxType_CanonicalExtension &&
        type != kUpb_TaggedAuxType_NonCanonicalExtension) {
      continue;
    }
    const upb_Extension* ext = aux.extension;
    const upb_MiniTableExtension* e = ext->ext;
    const upb_MiniTableField* f = &e->UPB_PRIVATE(field);
    const upb_MiniTable* m2 = upb_MiniTableExtension_GetSubMessage(e);

    upb_MessageValue val;
    memcpy(&val, &(ext->data), sizeof(upb_MessageValue));

    switch (UPB_PRIVATE(_upb_MiniTableField_Mode)(f)) {
      case kUpb_FieldMode_Array: {
        upb_Array* arr = (upb_Array*)val.array_val;
        if (arr) upb_Array_Freeze(arr, m2);
        break;
      }
      case kUpb_FieldMode_Map:
        UPB_UNREACHABLE();  // Maps cannot be extensions.
        break;
      case kUpb_FieldMode_Scalar:
        if (upb_MiniTableField_IsSubMessage(f)) {
          upb_Message* msg2 = (upb_Message*)val.msg_val;
          if (msg2) upb_Message_Freeze(msg2, m2);
        }
        break;
    }
  }
}

#include "stddef.h"

// Must be last.

bool upb_Message_MergeFrom(upb_Message* dst, const upb_Message* src,
                           const upb_MiniTable* mt,
                           const upb_ExtensionRegistry* extreg,
                           upb_Arena* arena) {
  char* buf = NULL;
  size_t size = 0;
  // This tmp arena is used to hold the bytes for `src` serialized. This bends
  // the typical "no hidden allocations" design of upb, but under a properly
  // optimized implementation this extra allocation would not be necessary and
  // so we don't want to unnecessarily have the bad API or bloat the passed-in
  // arena with this very-short-term allocation.
  upb_Arena* encode_arena = upb_Arena_New();
  upb_EncodeStatus e_status = upb_Encode(src, mt, 0, encode_arena, &buf, &size);
  if (e_status != kUpb_EncodeStatus_Ok) {
    upb_Arena_Free(encode_arena);
    return false;
  }
  upb_DecodeStatus d_status = upb_Decode(buf, size, dst, mt, extreg, 0, arena);
  if (d_status != kUpb_DecodeStatus_Ok) {
    upb_Arena_Free(encode_arena);
    return false;
  }
  upb_Arena_Free(encode_arena);
  return true;
}


#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Our awkward dance for including fasttable only when it is enabled.
#if UPB_FASTTABLE
#define UPB_INCLUDE_FAST_DECODE
#endif

#ifdef UPB_INCLUDE_FAST_DECODE
#endif

#undef UPB_INCLUDE_FAST_DECODE

// Must be last.

bool upb_MiniTable_SetSubMessage(upb_MiniTable* table,
                                 upb_MiniTableField* field,
                                 const upb_MiniTable* sub) {
  UPB_ASSERT((uintptr_t)table->UPB_PRIVATE(fields) <= (uintptr_t)field &&
             (uintptr_t)field < (uintptr_t)(table->UPB_PRIVATE(fields) +
                                            table->UPB_PRIVATE(field_count)));
  UPB_ASSERT(sub);

  const bool sub_is_map = sub->UPB_PRIVATE(ext) & kUpb_ExtMode_IsMapEntry;

  switch (field->UPB_PRIVATE(descriptortype)) {
    case kUpb_FieldType_Message:
      if (sub_is_map) {
        const bool table_is_map =
            table->UPB_PRIVATE(ext) & kUpb_ExtMode_IsMapEntry;
        if (UPB_UNLIKELY(table_is_map)) return false;

        field->UPB_PRIVATE(mode) =
            (field->UPB_PRIVATE(mode) & ~kUpb_FieldMode_Mask) |
            kUpb_FieldMode_Map;

#if UPB_FASTTABLE
        // The fasttable decoder cannot decode maps. Unfortunately we do not
        // know until this moment that the field is a map, so we have to
        // overwrite the fasttable entry (if any) that we built for this field
        // previously.
        int size = table->UPB_PRIVATE(table_mask) == 0xff
                       ? 0
                       : ((table->UPB_PRIVATE(table_mask) >> 3) + 1);
        for (int i = 0; i < size; i++) {
          _upb_FastTable_Entry* entry = &table->UPB_PRIVATE(fasttable)[i];
          uint32_t field_number = (((int)entry->field_data >> 3) & 0xf) |
                                  (((int)entry->field_data >> 4) & 0x7f0);
          if (field_number == upb_MiniTableField_Number(field)) {
            entry->field_parser = &_upb_FastDecoder_DecodeGeneric;
            entry->field_data = 0;
          }
        }
#endif
      }
      break;

    case kUpb_FieldType_Group:
      if (UPB_UNLIKELY(sub_is_map)) return false;
      break;

    default:
      return false;
  }

  upb_MiniTableSubInternal* table_sub =
      UPB_PTR_AT(field, field->UPB_PRIVATE(submsg_ofs) * kUpb_SubmsgOffsetBytes,
                 upb_MiniTableSubInternal);
  // TODO: Add this assert back once YouTube is updated to not call
  // this function repeatedly.
  // UPB_ASSERT(upb_MiniTable_GetSubMessageTable(table, field) == NULL);
  table_sub->UPB_PRIVATE(submsg) = sub;
  return true;
}

bool upb_MiniTable_SetSubEnum(upb_MiniTable* table, upb_MiniTableField* field,
                              const upb_MiniTableEnum* sub) {
  UPB_ASSERT((uintptr_t)table->UPB_PRIVATE(fields) <= (uintptr_t)field &&
             (uintptr_t)field < (uintptr_t)(table->UPB_PRIVATE(fields) +
                                            table->UPB_PRIVATE(field_count)));
  UPB_ASSERT(sub);

  if (field->UPB_PRIVATE(descriptortype) != kUpb_FieldType_Enum) {
    return false;
  }

  if ((table->UPB_PRIVATE(ext) & kUpb_ExtMode_IsMapEntry) &&
      !upb_MiniTableEnum_CheckValue(sub, 0)) {
    // An enum used in a map must include 0 as a value.  This matches a check
    // performed in protoc ("Enum value in map must define 0 as the first
    // value").  Protoc should ensure that we never get here.
    //
    // This ends up being important if we receive wire messages where a map
    // entry omits the value (and thus defaults to 0).
    return false;
  }

  upb_MiniTableSubInternal* table_sub =
      UPB_PTR_AT(field, field->UPB_PRIVATE(submsg_ofs) * kUpb_SubmsgOffsetBytes,
                 upb_MiniTableSubInternal);
  *table_sub = upb_MiniTableSub_FromEnum(sub);
  return true;
}

uint32_t upb_MiniTable_GetSubList(const upb_MiniTable* m,
                                  const upb_MiniTableField** subs) {
  uint32_t msg_count = 0;
  uint32_t enum_count = 0;

  for (int i = 0; i < upb_MiniTable_FieldCount(m); i++) {
    const upb_MiniTableField* f = upb_MiniTable_GetFieldByIndex(m, i);
    if (upb_MiniTableField_CType(f) == kUpb_CType_Message) {
      *subs = f;
      ++subs;
      msg_count++;
    }
  }

  for (int i = 0; i < upb_MiniTable_FieldCount(m); i++) {
    const upb_MiniTableField* f = upb_MiniTable_GetFieldByIndex(m, i);
    if (upb_MiniTableField_IsClosedEnum(f)) {
      *subs = f;
      ++subs;
      enum_count++;
    }
  }

  return (msg_count << 16) | enum_count;
}

// The list of sub_tables and sub_enums must exactly match the number and order
// of sub-message fields and sub-enum fields given by upb_MiniTable_GetSubList()
// above.
bool upb_MiniTable_Link(upb_MiniTable* m, const upb_MiniTable** sub_tables,
                        size_t sub_table_count,
                        const upb_MiniTableEnum** sub_enums,
                        size_t sub_enum_count) {
  uint32_t msg_count = 0;
  uint32_t enum_count = 0;

  for (int i = 0; i < upb_MiniTable_FieldCount(m); i++) {
    upb_MiniTableField* f =
        (upb_MiniTableField*)upb_MiniTable_GetFieldByIndex(m, i);
    if (upb_MiniTableField_CType(f) == kUpb_CType_Message) {
      const upb_MiniTable* sub = sub_tables[msg_count++];
      if (msg_count > sub_table_count) return false;
      if (sub && !upb_MiniTable_SetSubMessage(m, f, sub)) return false;
    }
  }

  for (int i = 0; i < upb_MiniTable_FieldCount(m); i++) {
    upb_MiniTableField* f =
        (upb_MiniTableField*)upb_MiniTable_GetFieldByIndex(m, i);
    if (upb_MiniTableField_IsClosedEnum(f)) {
      const upb_MiniTableEnum* sub = sub_enums[enum_count++];
      if (enum_count > sub_enum_count) return false;
      if (sub && !upb_MiniTable_SetSubEnum(m, f, sub)) return false;
    }
  }

  return (msg_count == sub_table_count) && (enum_count == sub_enum_count);
}


#include <inttypes.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>


// Our awkward dance for including fasttable only when it is enabled.
#if UPB_FASTTABLE
#define UPB_INCLUDE_FAST_DECODE
#endif

#ifdef UPB_INCLUDE_FAST_DECODE
#endif

#undef UPB_INCLUDE_FAST_DECODE

// Must be last.

// We reserve unused hasbits to make room for upb_Message fields.
#define kUpb_Reserved_Hasbytes sizeof(struct upb_Message)

// 64 is the first hasbit that we currently use.
#define kUpb_Reserved_Hasbits (kUpb_Reserved_Hasbytes * 8)

#define kUpb_OneOfLayoutItem_IndexSentinel ((uint16_t)-1)

// Stores the field number of the present value of the oneof
#define kUpb_OneOf_CaseFieldRep (kUpb_FieldRep_4Byte)

// The maximum field number that can be encoded on the wire.
// Note that this limit does not apply to MessageSet, which can have field
// numbers up to INT32_MAX.
#define kUpb_MaxFieldNumber ((1 << 29) - 1)

typedef struct {
  // Index of the corresponding field. The field's offset will be the index of
  // the next field in a linked list.
  uint16_t field_index;
  // This enum is stored in bytes to avoid trailing padding while preserving
  // two-byte alignment.
  uint8_t /* upb_FieldRep*/ rep;
} upb_OneOfLayoutItem;

typedef struct {
  upb_OneOfLayoutItem* data;
  size_t size;
  size_t buf_capacity_bytes;
} upb_OneOfLayoutItemVector;

typedef struct {
  upb_MdDecoder base;
  upb_MiniTableField* fields;
  upb_MiniTablePlatform platform;
  upb_OneOfLayoutItemVector oneofs;
  upb_Arena* arena;
  // Initially tracks the count of each field rep type; then, during assignment,
  // tracks the base offset for the next processed field of the given rep.
  uint16_t rep_counts_offsets[kUpb_FieldRep_Max + 1];
  uint32_t sub_count;
  bool is_extension;

  // When building fasttables, we don't know the total size of the table until
  // we're done decoding, so we allocate a temporary table on the stack and
  // allocate the real thing from the arena later.
  upb_MiniTable table;
} upb_MtDecoder;

// In each field's offset, we temporarily store a presence classifier:
enum PresenceClass {
  kNoPresence = 0,
  kHasbitPresence = 1,
  kRequiredPresence = 2,
  kOneofBase = 3,
  // Negative values refer to a specific oneof with that number.  Positive
  // values >= kOneofBase indicate that this field is in a oneof, and specify
  // the next field in this oneof's linked list.
};

static bool upb_MtDecoder_FieldIsPackable(upb_MiniTableField* field) {
  return (field->UPB_PRIVATE(mode) & kUpb_FieldMode_Array) &&
         upb_FieldType_IsPackable(field->UPB_PRIVATE(descriptortype));
}

typedef struct {
  uint16_t submsg_count;
  uint16_t subenum_count;
} upb_SubCounts;

static void upb_MiniTable_SetTypeAndSub(upb_MtDecoder* d,
                                        upb_MiniTableField* field,
                                        upb_FieldType type,
                                        uint64_t msg_modifiers,
                                        bool is_proto3_enum) {
  if (is_proto3_enum) {
    UPB_ASSERT(type == kUpb_FieldType_Enum);
    type = kUpb_FieldType_Int32;
    field->UPB_PRIVATE(mode) |= kUpb_LabelFlags_IsAlternate;
  } else if (type == kUpb_FieldType_String &&
             !(msg_modifiers & kUpb_MessageModifier_ValidateUtf8)) {
    type = kUpb_FieldType_Bytes;
    field->UPB_PRIVATE(mode) |= kUpb_LabelFlags_IsAlternate;
  }

  field->UPB_PRIVATE(descriptortype) = type;

  if (upb_MtDecoder_FieldIsPackable(field) &&
      (msg_modifiers & kUpb_MessageModifier_DefaultIsPacked)) {
    field->UPB_PRIVATE(mode) |= kUpb_LabelFlags_IsPacked;
  }

  // We initially set `submsg_ofs` to the index of the sub in the list of subs.
  // Later, we'll update it to be a relative byte offset.
  if (type == kUpb_FieldType_Message || type == kUpb_FieldType_Group ||
      type == kUpb_FieldType_Enum) {
    field->UPB_PRIVATE(submsg_ofs) = d->sub_count++;
  } else {
    field->UPB_PRIVATE(submsg_ofs) = kUpb_NoSub;
  }
}

static const char kUpb_EncodedToType[] = {
    [kUpb_EncodedType_Double] = kUpb_FieldType_Double,
    [kUpb_EncodedType_Float] = kUpb_FieldType_Float,
    [kUpb_EncodedType_Int64] = kUpb_FieldType_Int64,
    [kUpb_EncodedType_UInt64] = kUpb_FieldType_UInt64,
    [kUpb_EncodedType_Int32] = kUpb_FieldType_Int32,
    [kUpb_EncodedType_Fixed64] = kUpb_FieldType_Fixed64,
    [kUpb_EncodedType_Fixed32] = kUpb_FieldType_Fixed32,
    [kUpb_EncodedType_Bool] = kUpb_FieldType_Bool,
    [kUpb_EncodedType_String] = kUpb_FieldType_String,
    [kUpb_EncodedType_Group] = kUpb_FieldType_Group,
    [kUpb_EncodedType_Message] = kUpb_FieldType_Message,
    [kUpb_EncodedType_Bytes] = kUpb_FieldType_Bytes,
    [kUpb_EncodedType_UInt32] = kUpb_FieldType_UInt32,
    [kUpb_EncodedType_OpenEnum] = kUpb_FieldType_Enum,
    [kUpb_EncodedType_SFixed32] = kUpb_FieldType_SFixed32,
    [kUpb_EncodedType_SFixed64] = kUpb_FieldType_SFixed64,
    [kUpb_EncodedType_SInt32] = kUpb_FieldType_SInt32,
    [kUpb_EncodedType_SInt64] = kUpb_FieldType_SInt64,
    [kUpb_EncodedType_ClosedEnum] = kUpb_FieldType_Enum,
};

static void upb_MiniTable_SetField(upb_MtDecoder* d, uint8_t ch,
                                   upb_MiniTableField* field,
                                   uint64_t msg_modifiers) {
  static const char kUpb_EncodedToFieldRep[] = {
      [kUpb_EncodedType_Double] = kUpb_FieldRep_8Byte,
      [kUpb_EncodedType_Float] = kUpb_FieldRep_4Byte,
      [kUpb_EncodedType_Int64] = kUpb_FieldRep_8Byte,
      [kUpb_EncodedType_UInt64] = kUpb_FieldRep_8Byte,
      [kUpb_EncodedType_Int32] = kUpb_FieldRep_4Byte,
      [kUpb_EncodedType_Fixed64] = kUpb_FieldRep_8Byte,
      [kUpb_EncodedType_Fixed32] = kUpb_FieldRep_4Byte,
      [kUpb_EncodedType_Bool] = kUpb_FieldRep_1Byte,
      [kUpb_EncodedType_String] = kUpb_FieldRep_StringView,
      [kUpb_EncodedType_Bytes] = kUpb_FieldRep_StringView,
      [kUpb_EncodedType_UInt32] = kUpb_FieldRep_4Byte,
      [kUpb_EncodedType_OpenEnum] = kUpb_FieldRep_4Byte,
      [kUpb_EncodedType_SFixed32] = kUpb_FieldRep_4Byte,
      [kUpb_EncodedType_SFixed64] = kUpb_FieldRep_8Byte,
      [kUpb_EncodedType_SInt32] = kUpb_FieldRep_4Byte,
      [kUpb_EncodedType_SInt64] = kUpb_FieldRep_8Byte,
      [kUpb_EncodedType_ClosedEnum] = kUpb_FieldRep_4Byte,
  };

  char pointer_rep = d->platform == kUpb_MiniTablePlatform_32Bit
                         ? kUpb_FieldRep_4Byte
                         : kUpb_FieldRep_8Byte;

  int8_t type = _upb_FromBase92(ch);
  if (ch >= _upb_ToBase92(kUpb_EncodedType_RepeatedBase)) {
    type -= kUpb_EncodedType_RepeatedBase;
    field->UPB_PRIVATE(mode) = kUpb_FieldMode_Array;
    field->UPB_PRIVATE(mode) |= pointer_rep << kUpb_FieldRep_Shift;
    field->UPB_PRIVATE(offset) = kNoPresence;
  } else {
    field->UPB_PRIVATE(mode) = kUpb_FieldMode_Scalar;
    field->UPB_PRIVATE(offset) = kHasbitPresence;
    if (type == kUpb_EncodedType_Group || type == kUpb_EncodedType_Message) {
      field->UPB_PRIVATE(mode) |= pointer_rep << kUpb_FieldRep_Shift;
    } else if ((unsigned long)type >= sizeof(kUpb_EncodedToFieldRep)) {
      upb_MdDecoder_ErrorJmp(&d->base, "Invalid field type: %d", (int)type);
    } else {
      field->UPB_PRIVATE(mode) |= kUpb_EncodedToFieldRep[type]
                                  << kUpb_FieldRep_Shift;
    }
  }
  if ((unsigned long)type >= sizeof(kUpb_EncodedToType)) {
    upb_MdDecoder_ErrorJmp(&d->base, "Invalid field type: %d", (int)type);
  }
  upb_MiniTable_SetTypeAndSub(d, field, kUpb_EncodedToType[type], msg_modifiers,
                              type == kUpb_EncodedType_OpenEnum);
}

static void upb_MtDecoder_ModifyField(upb_MtDecoder* d,
                                      uint32_t message_modifiers,
                                      uint32_t field_modifiers,
                                      upb_MiniTableField* field) {
  if (field_modifiers & kUpb_EncodedFieldModifier_FlipPacked) {
    if (!upb_MtDecoder_FieldIsPackable(field)) {
      upb_MdDecoder_ErrorJmp(&d->base,
                             "Cannot flip packed on unpackable field %" PRIu32,
                             upb_MiniTableField_Number(field));
    }
    field->UPB_PRIVATE(mode) ^= kUpb_LabelFlags_IsPacked;
  }

  if (field_modifiers & kUpb_EncodedFieldModifier_FlipValidateUtf8) {
    if (field->UPB_PRIVATE(descriptortype) != kUpb_FieldType_Bytes ||
        !(field->UPB_PRIVATE(mode) & kUpb_LabelFlags_IsAlternate)) {
      upb_MdDecoder_ErrorJmp(&d->base,
                             "Cannot flip ValidateUtf8 on field %" PRIu32
                             ", type=%d, mode=%d",
                             upb_MiniTableField_Number(field),
                             (int)field->UPB_PRIVATE(descriptortype),
                             (int)field->UPB_PRIVATE(mode));
    }
    field->UPB_PRIVATE(descriptortype) = kUpb_FieldType_String;
    field->UPB_PRIVATE(mode) &= ~kUpb_LabelFlags_IsAlternate;
  }

  bool singular = field_modifiers & kUpb_EncodedFieldModifier_IsProto3Singular;
  bool required = field_modifiers & kUpb_EncodedFieldModifier_IsRequired;

  // Validate.
  if ((singular || required) && field->UPB_PRIVATE(offset) != kHasbitPresence) {
    upb_MdDecoder_ErrorJmp(&d->base,
                           "Invalid modifier(s) for repeated field %" PRIu32,
                           upb_MiniTableField_Number(field));
  }
  if (singular && required) {
    upb_MdDecoder_ErrorJmp(
        &d->base, "Field %" PRIu32 " cannot be both singular and required",
        upb_MiniTableField_Number(field));
  }

  if (singular && upb_MiniTableField_IsSubMessage(field)) {
    upb_MdDecoder_ErrorJmp(&d->base,
                           "Field %" PRIu32 " cannot be a singular submessage",
                           upb_MiniTableField_Number(field));
  }

  if (singular) field->UPB_PRIVATE(offset) = kNoPresence;
  if (required) {
    field->UPB_PRIVATE(offset) = kRequiredPresence;
  }
}

static void upb_MtDecoder_PushOneof(upb_MtDecoder* d,
                                    upb_OneOfLayoutItem item) {
  if (item.field_index == kUpb_OneOfLayoutItem_IndexSentinel) {
    upb_MdDecoder_ErrorJmp(&d->base, "Empty oneof");
  }
  if ((d->oneofs.size + 1) * sizeof(*d->oneofs.data) >
      d->oneofs.buf_capacity_bytes) {
    size_t new_cap = UPB_MAX(8, d->oneofs.size * 2) * sizeof(*d->oneofs.data);
    d->oneofs.data =
        upb_grealloc(d->oneofs.data, d->oneofs.buf_capacity_bytes, new_cap);
    upb_MdDecoder_CheckOutOfMemory(&d->base, d->oneofs.data);
    d->oneofs.buf_capacity_bytes = new_cap;
  }
  item.field_index -= kOneofBase;

  d->rep_counts_offsets[kUpb_OneOf_CaseFieldRep]++;
  d->rep_counts_offsets[item.rep]++;
  d->oneofs.data[d->oneofs.size++] = item;
}

static size_t upb_MtDecoder_SizeOfRep(upb_FieldRep rep,
                                      upb_MiniTablePlatform platform) {
  enum { string_view_size_32 = 8, string_view_size_64 = 16 };
  UPB_STATIC_ASSERT(sizeof(upb_StringView) ==
                        UPB_SIZE(string_view_size_32, string_view_size_64),
                    "StringView size mismatch");
  static const uint8_t kRepToSize32[] = {
      [kUpb_FieldRep_1Byte] = 1,
      [kUpb_FieldRep_4Byte] = 4,
      [kUpb_FieldRep_StringView] = string_view_size_32,
      [kUpb_FieldRep_8Byte] = 8,
  };
  static const uint8_t kRepToSize64[] = {
      [kUpb_FieldRep_1Byte] = 1,
      [kUpb_FieldRep_4Byte] = 4,
      [kUpb_FieldRep_StringView] = string_view_size_64,
      [kUpb_FieldRep_8Byte] = 8,
  };
  return platform == kUpb_MiniTablePlatform_32Bit ? kRepToSize32[rep]
                                                  : kRepToSize64[rep];
}

static size_t upb_MtDecoder_AlignOfRep(upb_FieldRep rep,
                                       upb_MiniTablePlatform platform) {
  enum { string_view_align_32 = 4, string_view_align_64 = 8 };
  UPB_STATIC_ASSERT(UPB_ALIGN_OF(upb_StringView) ==
                        UPB_SIZE(string_view_align_32, string_view_align_64),
                    "StringView size mismatch");

  static const uint8_t kRepToAlign32[] = {
      [kUpb_FieldRep_1Byte] = 1,
      [kUpb_FieldRep_4Byte] = 4,
      [kUpb_FieldRep_StringView] = string_view_align_32,
      [kUpb_FieldRep_8Byte] = 8,
  };
  static const uint8_t kRepToAlign64[] = {
      [kUpb_FieldRep_1Byte] = 1,
      [kUpb_FieldRep_4Byte] = 4,
      [kUpb_FieldRep_StringView] = string_view_align_64,
      [kUpb_FieldRep_8Byte] = 8,
  };
  return platform == kUpb_MiniTablePlatform_32Bit ? kRepToAlign32[rep]
                                                  : kRepToAlign64[rep];
}

static const char* upb_MtDecoder_DecodeOneofField(upb_MtDecoder* d,
                                                  const char* ptr,
                                                  char first_ch,
                                                  upb_OneOfLayoutItem* item) {
  uint32_t field_num;
  ptr = upb_MdDecoder_DecodeBase92Varint(
      &d->base, ptr, first_ch, kUpb_EncodedValue_MinOneofField,
      kUpb_EncodedValue_MaxOneofField, &field_num);
  upb_MiniTableField* f =
      (void*)upb_MiniTable_FindFieldByNumber(&d->table, field_num);

  if (!f) {
    upb_MdDecoder_ErrorJmp(&d->base,
                           "Couldn't add field number %" PRIu32
                           " to oneof, no such field number.",
                           field_num);
  }
  if (f->UPB_PRIVATE(offset) != kHasbitPresence) {
    upb_MdDecoder_ErrorJmp(
        &d->base,
        "Cannot add repeated, required, or singular field %" PRIu32
        " to oneof.",
        field_num);
  }

  // Oneof storage must be large enough to accommodate the largest member.
  int rep = f->UPB_PRIVATE(mode) >> kUpb_FieldRep_Shift;
  size_t new_size = upb_MtDecoder_SizeOfRep(rep, d->platform);
  size_t new_align = upb_MtDecoder_AlignOfRep(rep, d->platform);
  size_t current_size = upb_MtDecoder_SizeOfRep(item->rep, d->platform);
  size_t current_align = upb_MtDecoder_AlignOfRep(item->rep, d->platform);

  if (new_size > current_size ||
      (new_size == current_size && new_align > current_align)) {
    UPB_ASSERT(new_align >= current_align);
    item->rep = rep;
  } else {
    UPB_ASSERT(current_align >= new_align);
  }
  // Prepend this field to the linked list.
  f->UPB_PRIVATE(offset) = item->field_index;
  item->field_index = (f - d->fields) + kOneofBase;
  return ptr;
}

static const char* upb_MtDecoder_DecodeOneofs(upb_MtDecoder* d,
                                              const char* ptr) {
  upb_OneOfLayoutItem item = {
      .rep = 0, .field_index = kUpb_OneOfLayoutItem_IndexSentinel};
  while (ptr < d->base.end) {
    char ch = *ptr++;
    if (ch == kUpb_EncodedValue_FieldSeparator) {
      // Field separator, no action needed.
    } else if (ch == kUpb_EncodedValue_OneofSeparator) {
      // End of oneof.
      upb_MtDecoder_PushOneof(d, item);
      item.field_index =
          kUpb_OneOfLayoutItem_IndexSentinel;  // Move to next oneof.
    } else {
      ptr = upb_MtDecoder_DecodeOneofField(d, ptr, ch, &item);
    }
  }

  // Push final oneof.
  upb_MtDecoder_PushOneof(d, item);
  return ptr;
}

static const char* upb_MtDecoder_ParseModifier(upb_MtDecoder* d,
                                               const char* ptr, char first_ch,
                                               upb_MiniTableField* last_field,
                                               uint64_t* msg_modifiers) {
  uint32_t mod;
  ptr = upb_MdDecoder_DecodeBase92Varint(&d->base, ptr, first_ch,
                                         kUpb_EncodedValue_MinModifier,
                                         kUpb_EncodedValue_MaxModifier, &mod);
  if (last_field) {
    upb_MtDecoder_ModifyField(d, *msg_modifiers, mod, last_field);
  } else {
    if (d->is_extension) {
      upb_MdDecoder_ErrorJmp(&d->base,
                             "Extensions cannot have message modifiers");
    }
    *msg_modifiers = mod;
  }

  return ptr;
}

size_t upb_MtDecoder_PtrSize(upb_MtDecoder* d) {
  return d->platform == kUpb_MiniTablePlatform_32Bit ? 4 : 8;
}

static void upb_MtDecoder_AllocateSubs(upb_MtDecoder* d,
                                       upb_MiniTableSubInternal* subs) {
  // The `ofs` variable tracks byte offset between the current field and the
  // current entry in the `subs` array. Whenever we move to the next entry in
  // the `fields` array, the offset decreases by the size of the field, but
  // whenever we move to the next entry in the `subs` array, the offset
  // *increases* by the size of the entry in the `subs` array.
  UPB_ASSERT((char*)subs >= (char*)d->fields);
  size_t ofs = (char*)subs - (char*)d->fields;
  uintptr_t ptr_size = upb_MtDecoder_PtrSize(d);
  for (int i = 0; i < d->table.UPB_PRIVATE(field_count);
       i++, ofs -= sizeof(upb_MiniTableField)) {
    upb_MiniTableField* f = &d->fields[i];
    if (f->UPB_PRIVATE(submsg_ofs) == kUpb_NoSub) continue;
    size_t u32_ofs = ofs / kUpb_SubmsgOffsetBytes;
    UPB_ASSERT((ofs % 4) == 0);
    UPB_ASSERT((i * sizeof(upb_MiniTableField) + ofs) % ptr_size == 0);
    if (u32_ofs > UINT16_MAX) {
      upb_MdDecoder_ErrorJmp(&d->base, "Submessage offset overflow");
    }
    f->UPB_PRIVATE(submsg_ofs) = u32_ofs;
    ofs += ptr_size;
  }
}

static const char* upb_MtDecoder_Parse(upb_MtDecoder* d, const char* ptr,
                                       size_t len, void* fields,
                                       size_t field_size,
                                       uint16_t* field_count) {
  uint64_t msg_modifiers = 0;
  uint32_t last_field_number = 0;
  upb_MiniTableField* last_field = NULL;
  bool need_dense_below = !d->is_extension;

  d->base.end = UPB_PTRADD(ptr, len);

  while (ptr < d->base.end) {
    char ch = *ptr++;
    if (ch <= kUpb_EncodedValue_MaxField) {
      if (d->is_extension && last_field) {
        // For extensions, consume only a single field and then return.
        return --ptr;
      }
      upb_MiniTableField* field = fields;
      if (*field_count == UINT16_MAX) {
        upb_MdDecoder_ErrorJmp(
            &d->base, "Fields in message exceed the limit of %u", UINT16_MAX);
      }
      *field_count += 1;
      fields = (char*)fields + field_size;
      uint32_t number = ++last_field_number;
      if (number == 0 || (number > kUpb_MaxFieldNumber && !d->is_extension)) {
        upb_MdDecoder_ErrorJmp(&d->base, "Invalid field number: %" PRIu32,
                               number);
      }
      field->UPB_PRIVATE(number) = number;
      last_field = field;
      upb_MiniTable_SetField(d, ch, field, msg_modifiers);
    } else if (kUpb_EncodedValue_MinModifier <= ch &&
               ch <= kUpb_EncodedValue_MaxModifier) {
      ptr = upb_MtDecoder_ParseModifier(d, ptr, ch, last_field, &msg_modifiers);
      if (msg_modifiers & kUpb_MessageModifier_IsExtendable) {
        d->table.UPB_PRIVATE(ext) |= kUpb_ExtMode_Extendable;
      }
    } else if (ch == kUpb_EncodedValue_End) {
      if (d->is_extension) {
        upb_MdDecoder_ErrorJmp(&d->base, "Extensions cannot have oneofs.");
      }
      ptr = upb_MtDecoder_DecodeOneofs(d, ptr);
    } else if (kUpb_EncodedValue_MinSkip <= ch &&
               ch <= kUpb_EncodedValue_MaxSkip) {
      if (need_dense_below) {
        d->table.UPB_PRIVATE(dense_below) = d->table.UPB_PRIVATE(field_count);
        need_dense_below = false;
      }
      uint32_t skip;
      ptr = upb_MdDecoder_DecodeBase92Varint(&d->base, ptr, ch,
                                             kUpb_EncodedValue_MinSkip,
                                             kUpb_EncodedValue_MaxSkip, &skip);
      if (skip == 0) {
        upb_MdDecoder_ErrorJmp(&d->base, "Invalid skip value: 0");
      }
      last_field_number += skip;
      last_field_number--;  // Next field seen will increment.
    } else {
      upb_MdDecoder_ErrorJmp(&d->base, "Invalid char: %c", ch);
    }
  }

  if (need_dense_below) {
    d->table.UPB_PRIVATE(dense_below) = d->table.UPB_PRIVATE(field_count);
  }

  return ptr;
}

static void upb_MtDecoder_ParseMessage(upb_MtDecoder* d, const char* data,
                                       size_t len) {
  const size_t bytes_per_field =
      sizeof(upb_MiniTableField) + sizeof(upb_MiniTableSubInternal);
  // Buffer length is an upper bound on the number of fields. We will return
  // what we don't use.
  if ((SIZE_MAX - 4) / bytes_per_field < len) {
    upb_MdDecoder_ErrorJmp(&d->base, "MiniDescriptor is too large");
  }
  // Max size used per field is a upb_MiniTableField and a
  // upb_MiniTableSubInternal.  There could also be up to 4 bytes of padding,
  // since sizeof(upb_MiniTableField) == 12 and
  // alignof(upb_MiniTableSubInternal) == 8.
  UPB_STATIC_ASSERT(UPB_ALIGN_OF(upb_MiniTableSubInternal) -
                            UPB_ALIGN_OF(upb_MiniTableField) <=
                        4,
                    "alignment difference is too large");
  const size_t initial_bytes = bytes_per_field * len + 4;
  d->fields = upb_Arena_Malloc(d->arena, initial_bytes);
  upb_MdDecoder_CheckOutOfMemory(&d->base, d->fields);

  d->table.UPB_PRIVATE(field_count) = 0;
  d->table.UPB_PRIVATE(fields) = d->fields;
  upb_MtDecoder_Parse(d, data, len, d->fields, sizeof(*d->fields),
                      &d->table.UPB_PRIVATE(field_count));
  size_t field_bytes =
      UPB_ALIGN_UP(d->table.UPB_PRIVATE(field_count) * sizeof(*d->fields),
                   upb_MtDecoder_PtrSize(d));
  upb_MiniTableSubInternal* subs =
      UPB_PTR_AT(d->fields, field_bytes, upb_MiniTableSubInternal);
  memset(subs, 0, sizeof(upb_MiniTableSubInternal) * d->sub_count);

  // We now know how much space we actually used, so shrink the allocation to
  // that size.
  size_t final_bytes =
      field_bytes + sizeof(upb_MiniTableSubInternal) * d->sub_count;
  upb_Arena_ShrinkLast(d->arena, d->fields, initial_bytes, final_bytes);
  upb_MtDecoder_AllocateSubs(d, subs);
}

static void upb_MtDecoder_CalculateAlignments(upb_MtDecoder* d) {
  // Add alignment counts for non-oneof fields (oneofs were added already)
  int n = d->table.UPB_PRIVATE(field_count);
  for (int i = 0; i < n; i++) {
    upb_MiniTableField* f = &d->fields[i];
    if (f->UPB_PRIVATE(offset) >= kOneofBase) continue;
    d->rep_counts_offsets[f->UPB_PRIVATE(mode) >> kUpb_FieldRep_Shift]++;
  }

  // Reserve properly aligned space for each type of field representation
  // present in this message. When we iterate over the fields, they will obtain
  // their offset from within the region matching their alignment requirements.
  size_t base = d->table.UPB_PRIVATE(size);
  // Start with the lowest alignment requirement, going up, because:
  // 1. If there are presence bits, we won't be aligned to start, but adding
  //    some lower-alignment fields may get us closer without wasting space to
  //    padding.
  // 2. The allocator enforces 8 byte alignment, so moving intermediate padding
  //    to trailing padding doesn't save us anything.
  for (upb_FieldRep rep = kUpb_FieldRep_1Byte; rep <= kUpb_FieldRep_Max;
       rep++) {
    uint16_t count = d->rep_counts_offsets[rep];
    if (count) {
      base = UPB_ALIGN_UP(base, upb_MtDecoder_AlignOfRep(rep, d->platform));
      // This entry now tracks the base offset for this field representation
      // type, instead of the count
      d->rep_counts_offsets[rep] = base;
      base += upb_MtDecoder_SizeOfRep(rep, d->platform) * count;
    }
  }
  static const size_t max = UINT16_MAX;
  if (base > max) {
    upb_MdDecoder_ErrorJmp(
        &d->base, "Message size exceeded maximum size of %zu bytes", max);
  }
  d->table.UPB_PRIVATE(size) = (uint16_t)base;
}

static size_t upb_MiniTable_DivideRoundUp(size_t n, size_t d) {
  return (n + d - 1) / d;
}

static void upb_MtDecoder_AssignHasbits(upb_MtDecoder* d) {
  int n = d->table.UPB_PRIVATE(field_count);
  size_t last_hasbit = kUpb_Reserved_Hasbits - 1;

  // First assign required fields, which must have the lowest hasbits.
  for (int i = 0; i < n; i++) {
    upb_MiniTableField* field =
        (upb_MiniTableField*)&d->table.UPB_PRIVATE(fields)[i];
    if (field->UPB_PRIVATE(offset) == kRequiredPresence) {
      field->presence = ++last_hasbit;
    } else if (field->UPB_PRIVATE(offset) == kNoPresence) {
      field->presence = 0;
    }
  }
  if (last_hasbit >= kUpb_Reserved_Hasbits + 63) {
    upb_MdDecoder_ErrorJmp(&d->base, "Too many required fields");
  }

  d->table.UPB_PRIVATE(required_count) =
      last_hasbit - (kUpb_Reserved_Hasbits - 1);
  UPB_ASSERT(d->table.UPB_PRIVATE(required_count) < 64);

  // Next assign non-required hasbit fields.
  for (int i = 0; i < n; i++) {
    upb_MiniTableField* field =
        (upb_MiniTableField*)&d->table.UPB_PRIVATE(fields)[i];
    if (field->UPB_PRIVATE(offset) == kHasbitPresence) {
      field->presence = ++last_hasbit;
    }
  }

  d->table.UPB_PRIVATE(size) =
      last_hasbit ? upb_MiniTable_DivideRoundUp(last_hasbit + 1, 8) : 0;
}

static size_t upb_MtDecoder_Place(upb_MtDecoder* d, upb_FieldRep rep) {
  size_t size = upb_MtDecoder_SizeOfRep(rep, d->platform);
  size_t offset = d->rep_counts_offsets[rep];
  d->rep_counts_offsets[rep] += size;
  return offset;
}

static void upb_MtDecoder_AssignOffsets(upb_MtDecoder* d) {
  upb_MiniTableField* field_end =
      UPB_PTRADD(d->fields, d->table.UPB_PRIVATE(field_count));
  for (upb_MiniTableField* field = d->fields; field < field_end; field++) {
    if (field->UPB_PRIVATE(offset) >= kOneofBase) continue;
    field->UPB_PRIVATE(offset) =
        upb_MtDecoder_Place(d, field->UPB_PRIVATE(mode) >> kUpb_FieldRep_Shift);
  }

  upb_OneOfLayoutItem* oneof_end = UPB_PTRADD(d->oneofs.data, d->oneofs.size);

  for (upb_OneOfLayoutItem* item = d->oneofs.data; item < oneof_end; item++) {
    upb_MiniTableField* f = &d->fields[item->field_index];
    uint16_t case_offset = upb_MtDecoder_Place(d, kUpb_OneOf_CaseFieldRep);
    uint16_t data_offset = upb_MtDecoder_Place(d, item->rep);
    while (true) {
      f->presence = ~case_offset;
      uint16_t next_offset = f->UPB_PRIVATE(offset);
      f->UPB_PRIVATE(offset) = data_offset;
      if (next_offset == kUpb_OneOfLayoutItem_IndexSentinel) break;
      UPB_ASSERT(next_offset - kOneofBase < d->table.UPB_PRIVATE(field_count));
      f = &d->fields[next_offset - kOneofBase];
    }
  }

  // Since messages are always allocated on arenas, we can save repeatedly
  // realigning by doing alignment at minitable construction time. We don't want
  // to align to UPB_MALLOC_ALIGN because it can change with sanitizers, and if
  // we're generating code we don't want to calculate size differently depending
  // on the proto compiler's host or build configuration.
  d->table.UPB_PRIVATE(size) =
      UPB_ALIGN_UP(d->table.UPB_PRIVATE(size), kUpb_Message_Align);
}

static void upb_MtDecoder_ValidateEntryField(upb_MtDecoder* d,
                                             const upb_MiniTableField* f,
                                             uint32_t expected_num) {
  const char* name = expected_num == 1 ? "key" : "val";
  const uint32_t f_number = upb_MiniTableField_Number(f);
  if (f_number != expected_num) {
    upb_MdDecoder_ErrorJmp(&d->base,
                           "map %s did not have expected number (%d vs %d)",
                           name, expected_num, f_number);
  }

  if (!upb_MiniTableField_IsScalar(f)) {
    upb_MdDecoder_ErrorJmp(
        &d->base, "map %s cannot be repeated or map, or be in oneof", name);
  }

  uint32_t not_ok_types;
  if (expected_num == 1) {
    not_ok_types = (1 << kUpb_FieldType_Float) | (1 << kUpb_FieldType_Double) |
                   (1 << kUpb_FieldType_Message) | (1 << kUpb_FieldType_Group) |
                   (1 << kUpb_FieldType_Bytes) | (1 << kUpb_FieldType_Enum);
  } else {
    not_ok_types = 1 << kUpb_FieldType_Group;
  }

  if ((1 << upb_MiniTableField_Type(f)) & not_ok_types) {
    upb_MdDecoder_ErrorJmp(&d->base, "map %s cannot have type %d", name,
                           (int)f->UPB_PRIVATE(descriptortype));
  }
}

static void upb_MtDecoder_ParseMap(upb_MtDecoder* d, const char* data,
                                   size_t len) {
  upb_MtDecoder_ParseMessage(d, data, len);
  upb_MtDecoder_AssignHasbits(d);

  if (UPB_UNLIKELY(d->table.UPB_PRIVATE(field_count) != 2)) {
    upb_MdDecoder_ErrorJmp(&d->base, "%hu fields in map",
                           d->table.UPB_PRIVATE(field_count));
    UPB_UNREACHABLE();
  }

  if (d->oneofs.size != 0) {
    upb_MdDecoder_ErrorJmp(&d->base, "Map entry cannot have oneof");
  }

  upb_MtDecoder_ValidateEntryField(d, &d->table.UPB_PRIVATE(fields)[0], 1);
  upb_MtDecoder_ValidateEntryField(d, &d->table.UPB_PRIVATE(fields)[1], 2);

  d->fields[0].UPB_PRIVATE(offset) = offsetof(upb_MapEntry, k);
  d->fields[1].UPB_PRIVATE(offset) = offsetof(upb_MapEntry, v);
  d->table.UPB_PRIVATE(size) = sizeof(upb_MapEntry);

  // Map entries have a special bit set to signal it's a map entry, used in
  // upb_MiniTable_SetSubMessage() below.
  d->table.UPB_PRIVATE(ext) |= kUpb_ExtMode_IsMapEntry;
}

static void upb_MtDecoder_ParseMessageSet(upb_MtDecoder* d, const char* data,
                                          size_t len) {
  if (len > 0) {
    upb_MdDecoder_ErrorJmp(&d->base, "Invalid message set encode length: %zu",
                           len);
  }

  upb_MiniTable* ret = &d->table;
  ret->UPB_PRIVATE(size) = kUpb_Reserved_Hasbytes;
  ret->UPB_PRIVATE(field_count) = 0;
  ret->UPB_PRIVATE(ext) = kUpb_ExtMode_IsMessageSet;
  ret->UPB_PRIVATE(dense_below) = 0;
  ret->UPB_PRIVATE(table_mask) = -1;
  ret->UPB_PRIVATE(required_count) = 0;
}

static upb_MiniTable* upb_MtDecoder_DoBuildMiniTableWithBuf(
    upb_MtDecoder* decoder, const char* data, size_t len, void** buf,
    size_t* buf_size) {
  decoder->table.UPB_PRIVATE(size) = kUpb_Reserved_Hasbytes;
  decoder->table.UPB_PRIVATE(field_count) = 0;
  decoder->table.UPB_PRIVATE(ext) = kUpb_ExtMode_NonExtendable;
  decoder->table.UPB_PRIVATE(dense_below) = 0;
  decoder->table.UPB_PRIVATE(table_mask) = -1;
  decoder->table.UPB_PRIVATE(required_count) = 0;
#ifdef UPB_TRACING_ENABLED
  // MiniTables built from MiniDescriptors will not be able to vend the message
  // name unless it is explicitly set with upb_MiniTable_SetFullName().
  decoder->table.UPB_PRIVATE(full_name) = 0;
#endif

  // Strip off and verify the version tag.
  if (!len--) goto done;
  const char vers = *data++;

  switch (vers) {
    case kUpb_EncodedVersion_MapV1:
      upb_MtDecoder_ParseMap(decoder, data, len);
      break;

    case kUpb_EncodedVersion_MessageV1:
      upb_MtDecoder_ParseMessage(decoder, data, len);
      upb_MtDecoder_AssignHasbits(decoder);
      upb_MtDecoder_CalculateAlignments(decoder);
      upb_MtDecoder_AssignOffsets(decoder);
      break;

    case kUpb_EncodedVersion_MessageSetV1:
      upb_MtDecoder_ParseMessageSet(decoder, data, len);
      break;

    default:
      upb_MdDecoder_ErrorJmp(&decoder->base, "Invalid message version: %c",
                             vers);
  }

done:
  *buf = decoder->oneofs.data;
  *buf_size = decoder->oneofs.buf_capacity_bytes;
  size_t mt_size = sizeof(upb_MiniTable);

#if UPB_FASTTABLE
  upb_DecodeFast_TableEntry fasttable[32];
  int fasttable_size = upb_DecodeFast_BuildTable(&decoder->table, fasttable);
  mt_size += fasttable_size * sizeof(fasttable[0]);
#endif

  upb_MiniTable* ret = upb_Arena_Malloc(decoder->arena, mt_size);
  memcpy(ret, &decoder->table, sizeof(*ret));

#if UPB_FASTTABLE
  ret->UPB_PRIVATE(table_mask) = upb_DecodeFast_GetTableMask(fasttable_size);
  for (int i = 0; i < fasttable_size; i++) {
    ret->UPB_PRIVATE(fasttable)[i].field_data = fasttable[i].function_data;
    ret->UPB_PRIVATE(fasttable)[i].field_parser =
        upb_DecodeFast_GetFunctionPointer(fasttable[i].function_idx);
  }
#endif
  UPB_PRIVATE(upb_MiniTable_CheckInvariants)(ret);

#ifndef NDEBUG
  for (int i = 1; i < upb_MiniTable_FieldCount(ret); i++) {
    const upb_MiniTableField* f1 = upb_MiniTable_GetFieldByIndex(ret, i - 1);
    const upb_MiniTableField* f2 = upb_MiniTable_GetFieldByIndex(ret, i);
    UPB_ASSERT(upb_MiniTableField_Number(f2) > upb_MiniTableField_Number(f1));
  }
#endif
  return ret;
}

static upb_MiniTable* upb_MtDecoder_BuildMiniTableWithBuf(
    upb_MtDecoder* const decoder, const char* const data, const size_t len,
    void** const buf, size_t* const buf_size) {
  if (UPB_SETJMP(decoder->base.err) != 0) {
    *buf = decoder->oneofs.data;
    *buf_size = decoder->oneofs.buf_capacity_bytes;
    return NULL;
  }

  return upb_MtDecoder_DoBuildMiniTableWithBuf(decoder, data, len, buf,
                                               buf_size);
}

upb_MiniTable* upb_MiniTable_BuildWithBuf(const char* data, size_t len,
                                          upb_MiniTablePlatform platform,
                                          upb_Arena* arena, void** buf,
                                          size_t* buf_size,
                                          upb_Status* status) {
  upb_MtDecoder decoder = {
      .base = {.status = status},
      .platform = platform,
      .sub_count = 0,
      .is_extension = false,
      .oneofs =
          {
              .data = *buf,
              .buf_capacity_bytes = *buf_size,
              .size = 0,
          },
      .arena = arena,
  };

  return upb_MtDecoder_BuildMiniTableWithBuf(&decoder, data, len, buf,
                                             buf_size);
}

static const char* upb_MtDecoder_DoBuildMiniTableExtension(
    upb_MtDecoder* decoder, const char* data, size_t len,
    upb_MiniTableExtension* ext, const upb_MiniTable* extendee,
    upb_MiniTableSub sub) {
  if (!(extendee->UPB_PRIVATE(ext) &
        (kUpb_ExtMode_Extendable | kUpb_ExtMode_IsMessageSet))) {
    upb_MdDecoder_ErrorJmp(&decoder->base, "Extendee is not extendable");
  }

  // If the string is non-empty then it must begin with a version tag.
  if (len) {
    if (*data != kUpb_EncodedVersion_ExtensionV1) {
      upb_MdDecoder_ErrorJmp(&decoder->base, "Invalid ext version: %c", *data);
    }
    data++;
    len--;
  }

  uint16_t count = 0;
  const char* ret =
      upb_MtDecoder_Parse(decoder, data, len, ext, sizeof(*ext), &count);
  if (!ret || count != 1) return NULL;

  upb_MiniTableField* f = &ext->UPB_PRIVATE(field);
  uint32_t fieldnum = upb_MiniTableField_Number(f);

  const uint32_t kMaxFieldNumber = (1 << 29) - 1;
  if (fieldnum == 0 ||
      (fieldnum > kMaxFieldNumber && !upb_MiniTable_IsMessageSet(extendee))) {
    upb_MdDecoder_ErrorJmp(&decoder->base, "Invalid extension field number");
  }

  if (upb_MiniTable_FindFieldByNumber(extendee, fieldnum) != NULL) {
    upb_MdDecoder_ErrorJmp(&decoder->base,
                           "Extension overlaps with a known field");
  }

  f->UPB_PRIVATE(mode) |= kUpb_LabelFlags_IsExtension;
  f->UPB_PRIVATE(offset) = 0;
  f->presence = 0;

  // In upb_MiniTableExtension, the `sub` member is a pointer-sized member that
  // directly follows the `field` member.
  f->UPB_PRIVATE(submsg_ofs) =
      UPB_ALIGN_UP(sizeof(upb_MiniTableField), upb_MtDecoder_PtrSize(decoder)) /
      kUpb_SubmsgOffsetBytes;

  if (extendee->UPB_PRIVATE(ext) & kUpb_ExtMode_IsMessageSet) {
    // Extensions of MessageSet must be messages.
    if (!upb_MiniTableField_IsSubMessage(f)) return NULL;

    // Extensions of MessageSet must be non-repeating.
    if (upb_MiniTableField_IsArray(f)) return NULL;
  }

  ext->UPB_PRIVATE(extendee) = extendee;
  ext->UPB_PRIVATE(sub) = sub;

  return ret;
}

static const char* upb_MtDecoder_BuildMiniTableExtension(
    upb_MtDecoder* const decoder, const char* const data, const size_t len,
    upb_MiniTableExtension* const ext, const upb_MiniTable* const extendee,
    const upb_MiniTableSub sub) {
  if (UPB_SETJMP(decoder->base.err) != 0) return NULL;
  return upb_MtDecoder_DoBuildMiniTableExtension(decoder, data, len, ext,
                                                 extendee, sub);
}

const char* _upb_MiniTableExtension_Init(const char* data, size_t len,
                                         upb_MiniTableExtension* ext,
                                         const upb_MiniTable* extendee,
                                         upb_MiniTableSub sub,
                                         upb_MiniTablePlatform platform,
                                         upb_Status* status) {
  upb_MtDecoder decoder = {
      .base = {.status = status},
      .arena = NULL,
      .platform = platform,
      .sub_count = 0,
      .is_extension = true,
  };

  return upb_MtDecoder_BuildMiniTableExtension(&decoder, data, len, ext,
                                               extendee, sub);
}

upb_MiniTableExtension* _upb_MiniTableExtension_Build(
    const char* data, size_t len, const upb_MiniTable* extendee,
    upb_MiniTableSub sub, upb_MiniTablePlatform platform, upb_Arena* arena,
    upb_Status* status) {
  upb_MiniTableExtension* ext =
      upb_Arena_Malloc(arena, sizeof(upb_MiniTableExtension));
  if (UPB_UNLIKELY(!ext)) return NULL;

  const char* ptr = _upb_MiniTableExtension_Init(data, len, ext, extendee, sub,
                                                 platform, status);
  if (UPB_UNLIKELY(!ptr)) return NULL;

  return ext;
}

upb_MiniTable* _upb_MiniTable_Build(const char* data, size_t len,
                                    upb_MiniTablePlatform platform,
                                    upb_Arena* arena, upb_Status* status) {
  void* buf = NULL;
  size_t size = 0;
  upb_MiniTable* ret = upb_MiniTable_BuildWithBuf(data, len, platform, arena,
                                                  &buf, &size, status);
  upb_gfree(buf);
  return ret;
}


const char _kUpb_ToBase92[] = {
    ' ', '!', '#', '$', '%', '&', '(', ')', '*', '+', ',', '-', '.', '/',
    '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=',
    '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K',
    'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y',
    'Z', '[', ']', '^', '_', '`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
    'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
    'w', 'x', 'y', 'z', '{', '|', '}', '~',
};

const int8_t _kUpb_FromBase92[] = {
    0,  1,  -1, 2,  3,  4,  5,  -1, 6,  7,  8,  9,  10, 11, 12, 13, 14, 15, 16,
    17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35,
    36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54,
    55, 56, 57, -1, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72,
    73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91,
};


#include <assert.h>
#include <stddef.h>
#include <stdint.h>


// Must be last.

typedef struct {
  uint64_t present_values_mask;
  uint32_t last_written_value;
} upb_MtDataEncoderInternal_EnumState;

typedef struct {
  uint64_t msg_modifiers;
  uint32_t last_field_num;
  enum {
    kUpb_OneofState_NotStarted,
    kUpb_OneofState_StartedOneof,
    kUpb_OneofState_EmittedOneofField,
  } oneof_state;
} upb_MtDataEncoderInternal_MsgState;

typedef struct {
  char* buf_start;  // Only for checking kUpb_MtDataEncoder_MinSize.
  union {
    upb_MtDataEncoderInternal_EnumState enum_state;
    upb_MtDataEncoderInternal_MsgState msg_state;
  } state;
} upb_MtDataEncoderInternal;

static upb_MtDataEncoderInternal* upb_MtDataEncoder_GetInternal(
    upb_MtDataEncoder* e, char* buf_start) {
  UPB_ASSERT(sizeof(upb_MtDataEncoderInternal) <= sizeof(e->internal));
  upb_MtDataEncoderInternal* ret = (upb_MtDataEncoderInternal*)e->internal;
  ret->buf_start = buf_start;
  return ret;
}

static char* upb_MtDataEncoder_PutRaw(upb_MtDataEncoder* e, char* ptr,
                                      char ch) {
  upb_MtDataEncoderInternal* in = (upb_MtDataEncoderInternal*)e->internal;
  UPB_ASSERT(ptr - in->buf_start < kUpb_MtDataEncoder_MinSize);
  if (ptr == e->end) return NULL;
  *ptr++ = ch;
  return ptr;
}

static char* upb_MtDataEncoder_Put(upb_MtDataEncoder* e, char* ptr, char ch) {
  return upb_MtDataEncoder_PutRaw(e, ptr, _upb_ToBase92(ch));
}

static char* upb_MtDataEncoder_PutBase92Varint(upb_MtDataEncoder* e, char* ptr,
                                               uint32_t val, int min, int max) {
  int shift = upb_Log2Ceiling(_upb_FromBase92(max) - _upb_FromBase92(min) + 1);
  UPB_ASSERT(shift <= 6);
  uint32_t mask = (1 << shift) - 1;
  do {
    uint32_t bits = val & mask;
    ptr = upb_MtDataEncoder_Put(e, ptr, bits + _upb_FromBase92(min));
    if (!ptr) return NULL;
    val >>= shift;
  } while (val);
  return ptr;
}

char* upb_MtDataEncoder_PutModifier(upb_MtDataEncoder* e, char* ptr,
                                    uint64_t mod) {
  if (mod) {
    ptr = upb_MtDataEncoder_PutBase92Varint(e, ptr, mod,
                                            kUpb_EncodedValue_MinModifier,
                                            kUpb_EncodedValue_MaxModifier);
  }
  return ptr;
}

char* upb_MtDataEncoder_EncodeExtension(upb_MtDataEncoder* e, char* ptr,
                                        upb_FieldType type, uint32_t field_num,
                                        uint64_t field_mod) {
  upb_MtDataEncoderInternal* in = upb_MtDataEncoder_GetInternal(e, ptr);
  in->state.msg_state.msg_modifiers = 0;
  in->state.msg_state.last_field_num = 0;
  in->state.msg_state.oneof_state = kUpb_OneofState_NotStarted;

  ptr = upb_MtDataEncoder_PutRaw(e, ptr, kUpb_EncodedVersion_ExtensionV1);
  if (!ptr) return NULL;

  return upb_MtDataEncoder_PutField(e, ptr, type, field_num, field_mod);
}

char* upb_MtDataEncoder_EncodeMap(upb_MtDataEncoder* e, char* ptr,
                                  upb_FieldType key_type,
                                  upb_FieldType value_type, uint64_t key_mod,
                                  uint64_t value_mod) {
  upb_MtDataEncoderInternal* in = upb_MtDataEncoder_GetInternal(e, ptr);
  in->state.msg_state.msg_modifiers = 0;
  in->state.msg_state.last_field_num = 0;
  in->state.msg_state.oneof_state = kUpb_OneofState_NotStarted;

  ptr = upb_MtDataEncoder_PutRaw(e, ptr, kUpb_EncodedVersion_MapV1);
  if (!ptr) return NULL;

  ptr = upb_MtDataEncoder_PutField(e, ptr, key_type, 1, key_mod);
  if (!ptr) return NULL;

  return upb_MtDataEncoder_PutField(e, ptr, value_type, 2, value_mod);
}

char* upb_MtDataEncoder_EncodeMessageSet(upb_MtDataEncoder* e, char* ptr) {
  (void)upb_MtDataEncoder_GetInternal(e, ptr);
  return upb_MtDataEncoder_PutRaw(e, ptr, kUpb_EncodedVersion_MessageSetV1);
}

char* upb_MtDataEncoder_StartMessage(upb_MtDataEncoder* e, char* ptr,
                                     uint64_t msg_mod) {
  upb_MtDataEncoderInternal* in = upb_MtDataEncoder_GetInternal(e, ptr);
  in->state.msg_state.msg_modifiers = msg_mod;
  in->state.msg_state.last_field_num = 0;
  in->state.msg_state.oneof_state = kUpb_OneofState_NotStarted;

  ptr = upb_MtDataEncoder_PutRaw(e, ptr, kUpb_EncodedVersion_MessageV1);
  if (!ptr) return NULL;

  return upb_MtDataEncoder_PutModifier(e, ptr, msg_mod);
}

static char* _upb_MtDataEncoder_MaybePutFieldSkip(upb_MtDataEncoder* e,
                                                  char* ptr,
                                                  uint32_t field_num) {
  upb_MtDataEncoderInternal* in = (upb_MtDataEncoderInternal*)e->internal;
  if (field_num <= in->state.msg_state.last_field_num) return NULL;
  if (in->state.msg_state.last_field_num + 1 != field_num) {
    // Put skip.
    UPB_ASSERT(field_num > in->state.msg_state.last_field_num);
    uint32_t skip = field_num - in->state.msg_state.last_field_num;
    ptr = upb_MtDataEncoder_PutBase92Varint(
        e, ptr, skip, kUpb_EncodedValue_MinSkip, kUpb_EncodedValue_MaxSkip);
    if (!ptr) return NULL;
  }
  in->state.msg_state.last_field_num = field_num;
  return ptr;
}

static char* _upb_MtDataEncoder_PutFieldType(upb_MtDataEncoder* e, char* ptr,
                                             upb_FieldType type,
                                             uint64_t field_mod) {
  static const char kUpb_TypeToEncoded[] = {
      [kUpb_FieldType_Double] = kUpb_EncodedType_Double,
      [kUpb_FieldType_Float] = kUpb_EncodedType_Float,
      [kUpb_FieldType_Int64] = kUpb_EncodedType_Int64,
      [kUpb_FieldType_UInt64] = kUpb_EncodedType_UInt64,
      [kUpb_FieldType_Int32] = kUpb_EncodedType_Int32,
      [kUpb_FieldType_Fixed64] = kUpb_EncodedType_Fixed64,
      [kUpb_FieldType_Fixed32] = kUpb_EncodedType_Fixed32,
      [kUpb_FieldType_Bool] = kUpb_EncodedType_Bool,
      [kUpb_FieldType_String] = kUpb_EncodedType_String,
      [kUpb_FieldType_Group] = kUpb_EncodedType_Group,
      [kUpb_FieldType_Message] = kUpb_EncodedType_Message,
      [kUpb_FieldType_Bytes] = kUpb_EncodedType_Bytes,
      [kUpb_FieldType_UInt32] = kUpb_EncodedType_UInt32,
      [kUpb_FieldType_Enum] = kUpb_EncodedType_OpenEnum,
      [kUpb_FieldType_SFixed32] = kUpb_EncodedType_SFixed32,
      [kUpb_FieldType_SFixed64] = kUpb_EncodedType_SFixed64,
      [kUpb_FieldType_SInt32] = kUpb_EncodedType_SInt32,
      [kUpb_FieldType_SInt64] = kUpb_EncodedType_SInt64,
  };

  int encoded_type = kUpb_TypeToEncoded[type];

  if (field_mod & kUpb_FieldModifier_IsClosedEnum) {
    UPB_ASSERT(type == kUpb_FieldType_Enum);
    encoded_type = kUpb_EncodedType_ClosedEnum;
  }

  if (field_mod & kUpb_FieldModifier_IsRepeated) {
    // Repeated fields shift the type number up (unlike other modifiers which
    // are bit flags).
    encoded_type += kUpb_EncodedType_RepeatedBase;
  }

  return upb_MtDataEncoder_Put(e, ptr, encoded_type);
}

static char* _upb_MtDataEncoder_MaybePutModifiers(upb_MtDataEncoder* e,
                                                  char* ptr, upb_FieldType type,
                                                  uint64_t field_mod) {
  upb_MtDataEncoderInternal* in = (upb_MtDataEncoderInternal*)e->internal;
  uint32_t encoded_modifiers = 0;
  if ((field_mod & kUpb_FieldModifier_IsRepeated) &&
      upb_FieldType_IsPackable(type)) {
    bool field_is_packed = field_mod & kUpb_FieldModifier_IsPacked;
    bool default_is_packed = in->state.msg_state.msg_modifiers &
                             kUpb_MessageModifier_DefaultIsPacked;
    if (field_is_packed != default_is_packed) {
      encoded_modifiers |= kUpb_EncodedFieldModifier_FlipPacked;
    }
  }

  if (type == kUpb_FieldType_String) {
    bool field_validates_utf8 = field_mod & kUpb_FieldModifier_ValidateUtf8;
    bool message_validates_utf8 =
        in->state.msg_state.msg_modifiers & kUpb_MessageModifier_ValidateUtf8;
    if (field_validates_utf8 != message_validates_utf8) {
      // Old binaries do not recognize the field modifier.  We need the failure
      // mode to be too lax rather than too strict.  Our caller should have
      // handled this (see _upb_MessageDef_ValidateUtf8()).
      assert(!message_validates_utf8);
      encoded_modifiers |= kUpb_EncodedFieldModifier_FlipValidateUtf8;
    }
  }

  if (field_mod & kUpb_FieldModifier_IsProto3Singular) {
    encoded_modifiers |= kUpb_EncodedFieldModifier_IsProto3Singular;
  }

  if (field_mod & kUpb_FieldModifier_IsRequired) {
    encoded_modifiers |= kUpb_EncodedFieldModifier_IsRequired;
  }

  return upb_MtDataEncoder_PutModifier(e, ptr, encoded_modifiers);
}

char* upb_MtDataEncoder_PutField(upb_MtDataEncoder* e, char* ptr,
                                 upb_FieldType type, uint32_t field_num,
                                 uint64_t field_mod) {
  upb_MtDataEncoder_GetInternal(e, ptr);

  ptr = _upb_MtDataEncoder_MaybePutFieldSkip(e, ptr, field_num);
  if (!ptr) return NULL;

  ptr = _upb_MtDataEncoder_PutFieldType(e, ptr, type, field_mod);
  if (!ptr) return NULL;

  return _upb_MtDataEncoder_MaybePutModifiers(e, ptr, type, field_mod);
}

char* upb_MtDataEncoder_StartOneof(upb_MtDataEncoder* e, char* ptr) {
  upb_MtDataEncoderInternal* in = upb_MtDataEncoder_GetInternal(e, ptr);
  if (in->state.msg_state.oneof_state == kUpb_OneofState_NotStarted) {
    ptr = upb_MtDataEncoder_Put(e, ptr, _upb_FromBase92(kUpb_EncodedValue_End));
  } else {
    ptr = upb_MtDataEncoder_Put(
        e, ptr, _upb_FromBase92(kUpb_EncodedValue_OneofSeparator));
  }
  in->state.msg_state.oneof_state = kUpb_OneofState_StartedOneof;
  return ptr;
}

char* upb_MtDataEncoder_PutOneofField(upb_MtDataEncoder* e, char* ptr,
                                      uint32_t field_num) {
  upb_MtDataEncoderInternal* in = upb_MtDataEncoder_GetInternal(e, ptr);
  if (in->state.msg_state.oneof_state == kUpb_OneofState_EmittedOneofField) {
    ptr = upb_MtDataEncoder_Put(
        e, ptr, _upb_FromBase92(kUpb_EncodedValue_FieldSeparator));
    if (!ptr) return NULL;
  }
  ptr = upb_MtDataEncoder_PutBase92Varint(e, ptr, field_num, _upb_ToBase92(0),
                                          _upb_ToBase92(63));
  in->state.msg_state.oneof_state = kUpb_OneofState_EmittedOneofField;
  return ptr;
}

char* upb_MtDataEncoder_StartEnum(upb_MtDataEncoder* e, char* ptr) {
  upb_MtDataEncoderInternal* in = upb_MtDataEncoder_GetInternal(e, ptr);
  in->state.enum_state.present_values_mask = 0;
  in->state.enum_state.last_written_value = 0;

  return upb_MtDataEncoder_PutRaw(e, ptr, kUpb_EncodedVersion_EnumV1);
}

static char* upb_MtDataEncoder_FlushDenseEnumMask(upb_MtDataEncoder* e,
                                                  char* ptr) {
  upb_MtDataEncoderInternal* in = (upb_MtDataEncoderInternal*)e->internal;
  ptr = upb_MtDataEncoder_Put(e, ptr, in->state.enum_state.present_values_mask);
  in->state.enum_state.present_values_mask = 0;
  in->state.enum_state.last_written_value += 5;
  return ptr;
}

char* upb_MtDataEncoder_PutEnumValue(upb_MtDataEncoder* e, char* ptr,
                                     uint32_t val) {
  // TODO: optimize this encoding.
  upb_MtDataEncoderInternal* in = upb_MtDataEncoder_GetInternal(e, ptr);
  UPB_ASSERT(val >= in->state.enum_state.last_written_value);
  uint32_t delta = val - in->state.enum_state.last_written_value;
  if (delta >= 5 && in->state.enum_state.present_values_mask) {
    ptr = upb_MtDataEncoder_FlushDenseEnumMask(e, ptr);
    if (!ptr) {
      return NULL;
    }
    delta -= 5;
  }

  if (delta >= 5) {
    ptr = upb_MtDataEncoder_PutBase92Varint(
        e, ptr, delta, kUpb_EncodedValue_MinSkip, kUpb_EncodedValue_MaxSkip);
    in->state.enum_state.last_written_value += delta;
    delta = 0;
  }

  UPB_ASSERT((in->state.enum_state.present_values_mask >> delta) == 0);
  in->state.enum_state.present_values_mask |= 1ULL << delta;
  return ptr;
}

char* upb_MtDataEncoder_EndEnum(upb_MtDataEncoder* e, char* ptr) {
  upb_MtDataEncoderInternal* in = upb_MtDataEncoder_GetInternal(e, ptr);
  if (!in->state.enum_state.present_values_mask) return ptr;
  return upb_MtDataEncoder_FlushDenseEnumMask(e, ptr);
}


#include <stddef.h>
#include <stdint.h>


// Must be last.

typedef struct {
  upb_MdDecoder base;
  upb_Arena* arena;
  upb_MiniTableEnum* enum_table;
  uint32_t enum_value_count;
  uint32_t enum_data_count;
  uint32_t enum_data_capacity;
} upb_MdEnumDecoder;

static size_t upb_MiniTableEnum_Size(uint32_t count) {
  return UPB_SIZEOF_FLEX(upb_MiniTableEnum, UPB_PRIVATE(data), count);
}

static upb_MiniTableEnum* _upb_MiniTable_AddEnumDataMember(upb_MdEnumDecoder* d,
                                                           uint32_t val) {
  if (d->enum_data_count == d->enum_data_capacity) {
    size_t old_sz = upb_MiniTableEnum_Size(d->enum_data_capacity);
    if (d->enum_data_capacity > UINT32_MAX / 2) {
      upb_MdDecoder_ErrorJmp(&d->base, "Out of memory");
    }
    uint32_t new_capacity = UPB_MAX(2, d->enum_data_capacity * 2);
    if (UPB_SIZEOF_FLEX_WOULD_OVERFLOW(upb_MiniTableEnum, UPB_PRIVATE(data),
                                       new_capacity)) {
      upb_MdDecoder_ErrorJmp(&d->base, "Out of memory");
    }
    size_t new_sz = upb_MiniTableEnum_Size(new_capacity);
    d->enum_table = upb_Arena_Realloc(d->arena, d->enum_table, old_sz, new_sz);
    upb_MdDecoder_CheckOutOfMemory(&d->base, d->enum_table);
    d->enum_data_capacity = new_capacity;
  }
  d->enum_table->UPB_PRIVATE(data)[d->enum_data_count++] = val;
  return d->enum_table;
}

static void upb_MiniTableEnum_BuildValue(upb_MdEnumDecoder* d, uint32_t val) {
  upb_MiniTableEnum* table = d->enum_table;
  d->enum_value_count++;
  if (table->UPB_PRIVATE(value_count) ||
      (val > 512 && d->enum_value_count < val / 32)) {
    if (table->UPB_PRIVATE(value_count) == 0) {
      UPB_ASSERT(d->enum_data_count == table->UPB_PRIVATE(mask_limit) / 32);
    }
    table = _upb_MiniTable_AddEnumDataMember(d, val);
    table->UPB_PRIVATE(value_count)++;
  } else {
    uint32_t new_mask_limit = ((val / 32) + 1) * 32;
    while (table->UPB_PRIVATE(mask_limit) < new_mask_limit) {
      table = _upb_MiniTable_AddEnumDataMember(d, 0);
      table->UPB_PRIVATE(mask_limit) += 32;
    }
    table->UPB_PRIVATE(data)[val / 32] |= 1ULL << (val % 32);
  }
}

static upb_MiniTableEnum* upb_MtDecoder_DoBuildMiniTableEnum(
    upb_MdEnumDecoder* d, const char* data, size_t len) {
  // If the string is non-empty then it must begin with a version tag.
  if (len) {
    if (*data != kUpb_EncodedVersion_EnumV1) {
      upb_MdDecoder_ErrorJmp(&d->base, "Invalid enum version: %c", *data);
    }
    data++;
    len--;
  }

  upb_MdDecoder_CheckOutOfMemory(&d->base, d->enum_table);

  // Guarantee at least 64 bits of mask without checking mask size.
  d->enum_table->UPB_PRIVATE(mask_limit) = 64;
  d->enum_table = _upb_MiniTable_AddEnumDataMember(d, 0);
  d->enum_table = _upb_MiniTable_AddEnumDataMember(d, 0);

  d->enum_table->UPB_PRIVATE(value_count) = 0;

  const char* ptr = data;
  uint32_t base = 0;

  while (ptr < d->base.end) {
    char ch = *ptr++;
    if (ch <= kUpb_EncodedValue_MaxEnumMask) {
      uint32_t mask = _upb_FromBase92(ch);
      for (int i = 0; i < 5; i++, base++, mask >>= 1) {
        if (mask & 1) upb_MiniTableEnum_BuildValue(d, base);
      }
    } else if (kUpb_EncodedValue_MinSkip <= ch &&
               ch <= kUpb_EncodedValue_MaxSkip) {
      uint32_t skip;
      ptr = upb_MdDecoder_DecodeBase92Varint(&d->base, ptr, ch,
                                             kUpb_EncodedValue_MinSkip,
                                             kUpb_EncodedValue_MaxSkip, &skip);
      base += skip;
    } else {
      upb_MdDecoder_ErrorJmp(&d->base, "Unexpected character: %c", ch);
    }
  }

  return d->enum_table;
}

static upb_MiniTableEnum* upb_MtDecoder_BuildMiniTableEnum(
    upb_MdEnumDecoder* const decoder, const char* const data,
    size_t const len) {
  if (UPB_SETJMP(decoder->base.err) != 0) return NULL;
  return upb_MtDecoder_DoBuildMiniTableEnum(decoder, data, len);
}

upb_MiniTableEnum* upb_MiniTableEnum_Build(const char* data, size_t len,
                                           upb_Arena* arena,
                                           upb_Status* status) {
  uint32_t initial_capacity = 2;
  upb_MdEnumDecoder decoder = {
      .base =
          {
              .end = UPB_PTRADD(data, len),
              .status = status,
          },
      .arena = arena,
      .enum_table =
          upb_Arena_Malloc(arena, upb_MiniTableEnum_Size(initial_capacity)),
      .enum_value_count = 0,
      .enum_data_count = 0,
      .enum_data_capacity = initial_capacity,
  };

  return upb_MtDecoder_BuildMiniTableEnum(&decoder, data, len);
}


#include <inttypes.h>
#include <stdarg.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>


// Must be last.

typedef struct {
  char* buf;
  char* ptr;
  char* end;
  int overflow;
  upb_Arena* arena;
  int count;

  // This table maps from a pointer to a 64-bit integer. The lower 32 bits are
  // a unique ID for the object. The upper 32 bits are a flag that is 0x1
  // if the object has already been printed.
  upb_inttable inttable;
} upb_MiniTablePrinter;

UPB_PRINTF(2, 3)
static void upb_MiniTablePrinter_Printf(upb_MiniTablePrinter* p,
                                        const char* fmt, ...) {
  size_t n;
  size_t have = p->end - p->ptr;
  va_list args;

  va_start(args, fmt);
  n = _upb_vsnprintf(p->ptr, have, fmt, args);
  va_end(args);

  if (UPB_LIKELY(have > n)) {
    p->ptr += n;
  } else {
    p->ptr = UPB_PTRADD(p->ptr, have);
    p->overflow += (n - have);
  }
}

static size_t upb_MiniTablePrinter_NullTerminate(upb_MiniTablePrinter* p,
                                                 size_t size) {
  size_t ret = p->ptr - p->buf + p->overflow;

  if (size > 0) {
    if (p->ptr == p->end) p->ptr--;
    *p->ptr = '\0';
  }

  return ret;
}

static int upb_MiniTablePrinter_InsertNext(upb_MiniTablePrinter* p,
                                           const void* key, bool visited) {
  uint64_t id = p->count++;
  bool ok = upb_inttable_insert(
      &p->inttable, (intptr_t)key,
      upb_value_uint64(id | (visited ? 0x100000000 : 0)), p->arena);
  UPB_ASSERT(ok);
  return id;
}

// Returns the ID of the object referenced by key, but does *not* mark the
// object as visited. This is used for printing a reference to another object
// that may or may not have been printed yet.
static int upb_MiniTablePrinter_GetIdForRef(upb_MiniTablePrinter* p,
                                            const void* key) {
  upb_value v;
  if (upb_inttable_lookup(&p->inttable, (intptr_t)key, &v)) {
    return (int)upb_value_getuint64(v);
  }
  return upb_MiniTablePrinter_InsertNext(p, key, false);
}

// Returns the ID of the object referenced by key, and marks the object as
// visited. This is used for printing the object itself.
static int upb_MiniTablePrinter_GetIdForEmit(upb_MiniTablePrinter* p,
                                             const void* key) {
  UPB_ASSERT(key);
  upb_value v;
  if (upb_inttable_lookup(&p->inttable, (intptr_t)key, &v)) {
    uint64_t val = upb_value_getuint64(v);
    if (val >> 32) return -1;
    upb_inttable_replace(&p->inttable, (intptr_t)key,
                         upb_value_int64(val | 0x100000000));
    return (int)val;
  }
  return upb_MiniTablePrinter_InsertNext(p, key, true);
}

static void upb_MiniTablePrinter_PrintEnum(upb_MiniTablePrinter* p,
                                           const upb_MiniTableEnum* enum_) {
  int id = upb_MiniTablePrinter_GetIdForEmit(p, enum_);
  if (id < 0) return;

  upb_MiniTablePrinter_Printf(p, "MiniTableEnum#%d {\n", id);
  upb_MiniTablePrinter_Printf(p, "  .mask_limit = %d\n",
                              enum_->UPB_PRIVATE(mask_limit));
  upb_MiniTablePrinter_Printf(p, "  .value_count = %d\n",
                              enum_->UPB_PRIVATE(value_count));
  upb_MiniTablePrinter_Printf(p, "  .values = {\n");

  for (uint32_t i = 0; i < enum_->UPB_PRIVATE(mask_limit); i++) {
    if (!upb_MiniTableEnum_CheckValue(enum_, i)) continue;
    upb_MiniTablePrinter_Printf(p, "    %d,\n", (int)i);
  }

  const uint32_t* start =
      &enum_->UPB_PRIVATE(data)[enum_->UPB_PRIVATE(mask_limit) / 32];
  for (uint32_t i = 0; i < enum_->UPB_PRIVATE(value_count); i++) {
    upb_MiniTablePrinter_Printf(p, "    %d,\n", (int)start[i]);
  }

  upb_MiniTablePrinter_Printf(p, "  }\n");
  upb_MiniTablePrinter_Printf(p, "}\n\n");
}

static void upb_MiniTablePrinter_PrintField(upb_MiniTablePrinter* p,
                                            const upb_MiniTable* mini_table,
                                            const upb_MiniTableField* field) {
  upb_MiniTablePrinter_Printf(p, "    MiniTableField {\n");
  upb_MiniTablePrinter_Printf(p, "      .number = %d\n",
                              field->UPB_PRIVATE(number));
  upb_MiniTablePrinter_Printf(p, "      .offset = %d\n",
                              field->UPB_PRIVATE(offset));
  upb_MiniTablePrinter_Printf(p, "      .presence = %d", field->presence);

  if (field->presence > 0) {
    upb_MiniTablePrinter_Printf(p, " (hasbit=%d)\n", field->presence);
  } else if (field->presence < 0) {
    upb_MiniTablePrinter_Printf(p, " (oneof_index=%d)\n", ~field->presence);
  } else {
    upb_MiniTablePrinter_Printf(p, " (no explicit presence)\n");
  }

  if (field->UPB_PRIVATE(submsg_ofs) != kUpb_NoSub) {
    upb_MiniTablePrinter_Printf(p, "      .submsg_ofs = %d\n",
                                field->UPB_PRIVATE(submsg_ofs));
  }
  upb_MiniTablePrinter_Printf(p, "      .type = %d\n",
                              field->UPB_PRIVATE(descriptortype));
  upb_MiniTablePrinter_Printf(p, "      .mode = %02x (",
                              field->UPB_PRIVATE(mode));

  switch (field->UPB_PRIVATE(mode) & kUpb_FieldMode_Mask) {
    case kUpb_FieldMode_Scalar:
      upb_MiniTablePrinter_Printf(p, "Scalar");
      break;
    case kUpb_FieldMode_Array:
      upb_MiniTablePrinter_Printf(p, "Array");
      break;
    case kUpb_FieldMode_Map:
      upb_MiniTablePrinter_Printf(p, "Map");
      break;
  }

  switch (field->UPB_PRIVATE(mode) >> kUpb_FieldRep_Shift) {
    case kUpb_FieldRep_1Byte:
      upb_MiniTablePrinter_Printf(p, " | 1Byte");
      break;
    case kUpb_FieldRep_4Byte:
      upb_MiniTablePrinter_Printf(p, " | 4Byte");
      break;
    case kUpb_FieldRep_8Byte:
      upb_MiniTablePrinter_Printf(p, " | 8Byte");
      break;
    case kUpb_FieldRep_StringView:
      upb_MiniTablePrinter_Printf(p, " | StringView");
      break;
  }

  if (field->UPB_PRIVATE(mode) & kUpb_LabelFlags_IsPacked) {
    upb_MiniTablePrinter_Printf(p, " | Packed");
  }
  if (field->UPB_PRIVATE(mode) & kUpb_LabelFlags_IsExtension) {
    upb_MiniTablePrinter_Printf(p, " | Extension");
  }
  if (field->UPB_PRIVATE(mode) & kUpb_LabelFlags_IsAlternate) {
    upb_MiniTablePrinter_Printf(p, " | Alternate");
  }

  upb_MiniTablePrinter_Printf(p, ")\n");

  if (field->UPB_PRIVATE(submsg_ofs) != kUpb_NoSub) {
    if (upb_MiniTableField_CType(field) == kUpb_CType_Message) {
      int id =
          upb_MiniTablePrinter_GetIdForRef(p, upb_MiniTable_SubMessage(field));
      upb_MiniTablePrinter_Printf(p, "      .submsg = MiniTable#%d\n", id);
    } else {
      int id = upb_MiniTablePrinter_GetIdForRef(
          p, upb_MiniTable_GetSubEnumTable(field));
      upb_MiniTablePrinter_Printf(p, "      .subenum = MiniTableEnum#%d\n", id);
    }
  }

  upb_MiniTablePrinter_Printf(p, "    },\n");
}

static void upb_MiniTablePrinter_PrintMessage(upb_MiniTablePrinter* p,
                                              const upb_MiniTable* mini_table) {
  int id = upb_MiniTablePrinter_GetIdForEmit(p, mini_table);
  if (id < 0) return;

  upb_MiniTablePrinter_Printf(p, "MiniTable#%d {\n", id);
  upb_MiniTablePrinter_Printf(p, "  .size = %d\n",
                              mini_table->UPB_PRIVATE(size));
  upb_MiniTablePrinter_Printf(p, "  .required_count = %d\n",
                              mini_table->UPB_PRIVATE(required_count));
  upb_MiniTablePrinter_Printf(p, "  .table_mask = %02x\n",
                              mini_table->UPB_PRIVATE(table_mask));
  upb_MiniTablePrinter_Printf(p, "  .dense_below = %d\n",
                              mini_table->UPB_PRIVATE(dense_below));

  upb_MiniTablePrinter_Printf(p, "  .ext = %02x (",
                              mini_table->UPB_PRIVATE(ext));
  switch (mini_table->UPB_PRIVATE(ext) & 3) {
    case kUpb_ExtMode_NonExtendable:
      upb_MiniTablePrinter_Printf(p, "NonExtendable");
      break;
    case kUpb_ExtMode_Extendable:
      upb_MiniTablePrinter_Printf(p, "Extendable");
      break;
    case kUpb_ExtMode_IsMessageSet:
      upb_MiniTablePrinter_Printf(p, "MessageSet");
      break;
    case kUpb_ExtMode_IsMessageSet_ITEM:
      upb_MiniTablePrinter_Printf(p, "MessageSetItem");
      break;
  }
  if (mini_table->UPB_PRIVATE(ext) & kUpb_ExtMode_IsMapEntry) {
    upb_MiniTablePrinter_Printf(p, " | MapEntry");
  }
  if (mini_table->UPB_PRIVATE(ext) & kUpb_ExtMode_AllFastFieldsAssigned) {
    upb_MiniTablePrinter_Printf(p, " | AllFieldsAssigned");
  }
  upb_MiniTablePrinter_Printf(p, ")\n");
  upb_MiniTablePrinter_Printf(p, "  .fields[%d] = {\n",
                              mini_table->UPB_PRIVATE(field_count));

  for (int i = 0; i < mini_table->UPB_PRIVATE(field_count); i++) {
    const upb_MiniTableField* field = &mini_table->UPB_PRIVATE(fields)[i];
    upb_MiniTablePrinter_PrintField(p, mini_table, field);
  }

  upb_MiniTablePrinter_Printf(p, "  }\n");

#if UPB_FASTTABLE
  int mask = (int8_t)mini_table->UPB_PRIVATE(table_mask);
  if (mask != -1) {
    int size = (mask >> 3) + 1;
    upb_MiniTablePrinter_Printf(p, "  .fasttable[%d] = {\n", size);

    for (int i = 0; i < size; i++) {
      const _upb_FastTable_Entry* entry =
          &mini_table->UPB_PRIVATE(fasttable)[i];
      upb_MiniTablePrinter_Printf(p, "    FastTableEntry {\n");
      upb_MiniTablePrinter_Printf(p, "      .field_data = %016" PRIx64 ",\n",
                                  entry->field_data);
      upb_MiniTablePrinter_Printf(p, "      .field_parser = %s\n",
                                  entry->field_parser ? "[ptr]" : "(nil)");
      upb_MiniTablePrinter_Printf(p, "      .field_number = %d\n",
                                  (((int)entry->field_data >> 3) & 0xf) |
                                      (((int)entry->field_data >> 4) & 0x7f0));
      upb_MiniTablePrinter_Printf(p, "    }\n");
    }

    upb_MiniTablePrinter_Printf(p, "  }\n");
  }
#endif

  upb_MiniTablePrinter_Printf(p, "}\n\n");

  for (int i = 0; i < mini_table->UPB_PRIVATE(field_count); i++) {
    const upb_MiniTableField* field = &mini_table->UPB_PRIVATE(fields)[i];
    if (field->UPB_PRIVATE(submsg_ofs) == kUpb_NoSub) continue;
    if (upb_MiniTableField_CType(field) == kUpb_CType_Message) {
      upb_MiniTablePrinter_PrintMessage(p, upb_MiniTable_SubMessage(field));
    } else {
      upb_MiniTablePrinter_PrintEnum(p, upb_MiniTable_GetSubEnumTable(field));
    }
  }
}

size_t upb_MiniTable_DebugString(const upb_MiniTable* mini_table, char* buf,
                                 size_t size) {
  upb_MiniTablePrinter p = {buf, buf, buf + size, 0, upb_Arena_New(), 0};

  if (!p.arena) return 0;
  if (!upb_inttable_init(&p.inttable, p.arena)) return 0;

  upb_MiniTablePrinter_PrintMessage(&p, mini_table);

  upb_Arena_Free(p.arena);

  return upb_MiniTablePrinter_NullTerminate(&p, size);
}


#include <stddef.h>
#include <stdint.h>


// Must be last.

// Checks if source and target mini table fields are identical.
//
// If the field is a sub message and sub messages are identical we record
// the association in table.
//
// Hashing the source sub message mini table and it's equivalent in the table
// stops recursing when a cycle is detected and instead just checks if the
// destination table is equal.
static upb_MiniTableEquals_Status upb_deep_check(const upb_MiniTable* src,
                                                 const upb_MiniTable* dst,
                                                 upb_inttable* table,
                                                 upb_Arena** arena) {
  if (src->UPB_PRIVATE(field_count) != dst->UPB_PRIVATE(field_count))
    return kUpb_MiniTableEquals_NotEqual;
  bool marked_src = false;
  for (int i = 0; i < upb_MiniTable_FieldCount(src); i++) {
    const upb_MiniTableField* src_field = upb_MiniTable_GetFieldByIndex(src, i);
    const upb_MiniTableField* dst_field = upb_MiniTable_FindFieldByNumber(
        dst, upb_MiniTableField_Number(src_field));

    if (upb_MiniTableField_CType(src_field) !=
        upb_MiniTableField_CType(dst_field))
      return false;
    if (src_field->UPB_PRIVATE(mode) != dst_field->UPB_PRIVATE(mode))
      return false;
    if (src_field->UPB_PRIVATE(offset) != dst_field->UPB_PRIVATE(offset))
      return false;
    if (src_field->presence != dst_field->presence) return false;
    if (src_field->UPB_PRIVATE(submsg_ofs) !=
        dst_field->UPB_PRIVATE(submsg_ofs))
      return kUpb_MiniTableEquals_NotEqual;

    // Go no further if we are only checking for compatibility.
    if (!table) continue;

    if (upb_MiniTableField_CType(src_field) == kUpb_CType_Message) {
      if (!*arena) {
        *arena = upb_Arena_New();
        if (!upb_inttable_init(table, *arena)) {
          return kUpb_MiniTableEquals_OutOfMemory;
        }
      }
      if (!marked_src) {
        marked_src = true;
        upb_value val;
        val.val = (uint64_t)dst;
        if (!upb_inttable_insert(table, (uintptr_t)src, val, *arena)) {
          return kUpb_MiniTableEquals_OutOfMemory;
        }
      }
      const upb_MiniTable* sub_src =
          upb_MiniTable_GetSubMessageTable(src_field);
      const upb_MiniTable* sub_dst =
          upb_MiniTable_GetSubMessageTable(dst_field);
      if (sub_src != NULL) {
        upb_value cmp;
        if (upb_inttable_lookup(table, (uintptr_t)sub_src, &cmp)) {
          // We already compared this src before. Check if same dst.
          if (cmp.val != (uint64_t)sub_dst) {
            return kUpb_MiniTableEquals_NotEqual;
          }
        } else {
          // Recurse if not already visited.
          upb_MiniTableEquals_Status s =
              upb_deep_check(sub_src, sub_dst, table, arena);
          if (s != kUpb_MiniTableEquals_Equal) {
            return s;
          }
        }
      }
    }
  }
  return kUpb_MiniTableEquals_Equal;
}

bool upb_MiniTable_Compatible(const upb_MiniTable* src,
                              const upb_MiniTable* dst) {
  return upb_deep_check(src, dst, NULL, NULL);
}

upb_MiniTableEquals_Status upb_MiniTable_Equals(const upb_MiniTable* src,
                                                const upb_MiniTable* dst) {
  // Arena allocated on demand for hash table.
  upb_Arena* arena = NULL;
  // Table to keep track of visited mini tables to guard against cycles.
  upb_inttable table;
  upb_MiniTableEquals_Status status = upb_deep_check(src, dst, &table, &arena);
  if (arena) {
    upb_Arena_Free(arena);
  }
  return status;
}


#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Must be last.

struct upb_ExtensionRegistry {
  upb_exttable exts;
  upb_Arena* arena;
};

upb_ExtensionRegistry* upb_ExtensionRegistry_New(upb_Arena* arena) {
  upb_ExtensionRegistry* r = upb_Arena_Malloc(arena, sizeof(*r));
  if (!r) return NULL;
  r->arena = arena;
  if (!upb_exttable_init(&r->exts, 8, arena)) return NULL;
  return r;
}

UPB_API upb_ExtensionRegistryStatus upb_ExtensionRegistry_Add(
    upb_ExtensionRegistry* r, const upb_MiniTableExtension* e) {
  UPB_STATIC_ASSERT(
      offsetof(upb_MiniTableExtension,
               UPB_PRIVATE(field).UPB_PRIVATE(number)) == 0,
      "Extension must be first-member-of-struct convertable with uint32_t");
  uint32_t fieldnum = upb_MiniTableExtension_Number(e);
  const upb_MiniTable* extendee = upb_MiniTableExtension_Extendee(e);

  if (upb_exttable_lookup(&r->exts, extendee, fieldnum) != NULL) {
    return kUpb_ExtensionRegistryStatus_DuplicateEntry;
  }

  if (!upb_exttable_insert(&r->exts, extendee, (const uint32_t*)e, r->arena)) {
    return kUpb_ExtensionRegistryStatus_OutOfMemory;
  }
  return kUpb_ExtensionRegistryStatus_Ok;
}

upb_ExtensionRegistryStatus upb_ExtensionRegistry_AddArray(
    upb_ExtensionRegistry* r, const upb_MiniTableExtension** e, size_t count) {
  const upb_MiniTableExtension** start = e;
  const upb_MiniTableExtension** end = UPB_PTRADD(e, count);
  upb_ExtensionRegistryStatus status = kUpb_ExtensionRegistryStatus_Ok;
  for (; e < end; e++) {
    status = upb_ExtensionRegistry_Add(r, *e);
    if (status != kUpb_ExtensionRegistryStatus_Ok) goto failure;
  }
  return kUpb_ExtensionRegistryStatus_Ok;

failure:
  // Back out the entries previously added.
  for (end = e, e = start; e < end; e++) {
    const upb_MiniTableExtension* ext = *e;
    upb_exttable_remove(&r->exts, upb_MiniTableExtension_Extendee(ext),
                        upb_MiniTableExtension_Number(ext));
  }
  UPB_ASSERT(status != kUpb_ExtensionRegistryStatus_Ok);
  return status;
}

const upb_MiniTableExtension* upb_ExtensionRegistry_Lookup(
    const upb_ExtensionRegistry* r, const upb_MiniTable* t, uint32_t num) {
  const uint32_t* v = upb_exttable_lookup(&r->exts, t, num);
  return (const upb_MiniTableExtension*)v;
}

size_t upb_ExtensionRegistry_Size(const upb_ExtensionRegistry* r) {
  return upb_exttable_size(&r->exts);
}


#include <stddef.h>


// Must be last.

// A MiniTable for a statically tree shaken message.  Messages that use this
// MiniTable are guaranteed to remain unlinked; unlike the empty message, this
// MiniTable is never replaced, which greatly simplifies everything, because the
// type of a sub-message is always known, without consulting a tagged bit.
const struct upb_MiniTable UPB_PRIVATE(_kUpb_MiniTable_StaticallyTreeShaken) = {
    .UPB_PRIVATE(fields) = NULL,
    .UPB_PRIVATE(size) = sizeof(struct upb_Message),
    .UPB_PRIVATE(field_count) = 0,
    .UPB_PRIVATE(ext) = kUpb_ExtMode_NonExtendable,
    .UPB_PRIVATE(dense_below) = 0,
    .UPB_PRIVATE(table_mask) = -1,
    .UPB_PRIVATE(required_count) = 0,
};


#include <stddef.h>
#include <stdint.h>


// Must be last.

#if UPB_TSAN
#include <sched.h>
#endif  // UPB_TSAN

const UPB_PRIVATE(upb_GeneratedExtensionListEntry) *
    UPB_PRIVATE(upb_generated_extension_list) = NULL;

typedef struct upb_GeneratedRegistry {
  UPB_ATOMIC(upb_GeneratedRegistryRef*) ref;
  UPB_ATOMIC(int32_t) ref_count;
} upb_GeneratedRegistry;

static upb_GeneratedRegistry* _upb_generated_registry(void) {
  static upb_GeneratedRegistry r = {NULL, 0};
  return &r;
}

static bool _upb_GeneratedRegistry_AddAllLinkedExtensions(
    upb_ExtensionRegistry* r) {
  const UPB_PRIVATE(upb_GeneratedExtensionListEntry)* entry =
      UPB_PRIVATE(upb_generated_extension_list);
  while (entry != NULL) {
    const upb_MiniTableExtension** current = entry->start;
    for (current = entry->start; current != entry->stop; ++current) {
      const upb_MiniTableExtension* ext = *current;
      // Sentinels and padding introduced by the linker can result in zeroed
      // entries, so simply skip them.
      if (*current == NULL) {
        // MSVC introduces padding that might not be sized exactly the same as
        // the linker array element, but it should be properly aligned, so just
        // skipping empty elements should be safe.  (If the size and align of
        // the array elements was different, we'd have to do something more
        // complicated).
        continue;
      }

      if (upb_ExtensionRegistry_Add(r, ext) !=
          kUpb_ExtensionRegistryStatus_Ok) {
        return false;
      }
    }
    entry = entry->next;
  }
  return true;
}

// Constructs a new GeneratedRegistryRef, adding all linked extensions to the
// registry or returning NULL on failure.
static upb_GeneratedRegistryRef* _upb_GeneratedRegistry_New(void) {
  upb_Arena* arena = NULL;
  upb_ExtensionRegistry* extreg = NULL;
  upb_GeneratedRegistryRef* ref = upb_gmalloc(sizeof(upb_GeneratedRegistryRef));
  if (ref == NULL) goto err;
  arena = upb_Arena_New();
  if (arena == NULL) goto err;
  extreg = upb_ExtensionRegistry_New(arena);
  if (extreg == NULL) goto err;

  ref->UPB_PRIVATE(arena) = arena;
  ref->UPB_PRIVATE(registry) = extreg;

  if (!_upb_GeneratedRegistry_AddAllLinkedExtensions(extreg)) goto err;

  return ref;

err:
  if (arena != NULL) upb_Arena_Free(arena);
  if (ref != NULL) upb_gfree(ref);
  return NULL;
}

const upb_GeneratedRegistryRef* upb_GeneratedRegistry_Load(void) {
  upb_GeneratedRegistry* registry = _upb_generated_registry();

  // Loop until we successfully acquire a reference.  This loop should only
  // kick in under extremely high contention, and it should be guaranteed to
  // succeed.
  while (true) {
    int32_t count = upb_Atomic_Load(&registry->ref_count, memory_order_acquire);

    // Try to increment the refcount, but only if it's not zero.
    while (count > 0) {
      if (upb_Atomic_CompareExchangeStrong(&registry->ref_count, &count,
                                           count + 1, memory_order_acquire,
                                           memory_order_relaxed)) {
        // Successfully incremented. We can now safely load and return the
        // pointer.
        const upb_GeneratedRegistryRef* ref =
            upb_Atomic_Load(&registry->ref, memory_order_acquire);
        UPB_ASSERT(ref != NULL);
        return ref;
      }
      // CAS failed, `count` was updated. Loop will retry.
    }

    // If we're here, the count was 0. Time for the slow path.
    // Double-check that the pointer is NULL before trying to create.
    upb_GeneratedRegistryRef* ref =
        upb_Atomic_Load(&registry->ref, memory_order_acquire);
    if (ref == NULL) {
      // Pointer is NULL, try to create and publish a new registry.
      upb_GeneratedRegistryRef* new_ref = _upb_GeneratedRegistry_New();
      if (new_ref == NULL) return NULL;  // OOM

      // Try to CAS the pointer from NULL to our new_ref.
      if (upb_Atomic_CompareExchangeStrong(&registry->ref, &ref, new_ref,
                                           memory_order_release,
                                           memory_order_acquire)) {
        // We won the race. Set the ref count to 1.
        upb_Atomic_Store(&registry->ref_count, 1, memory_order_release);
        return new_ref;
      } else {
        // We lost the race. `ref` now holds the pointer from the winning
        // thread. Clean up our unused one and loop to try again to get a
        // reference.
        upb_Arena_Free(new_ref->UPB_PRIVATE(arena));
        upb_gfree(new_ref);
      }
    }
    // If we are here, either we lost the CAS race, or the pointer was already
    // non-NULL. In either case, we loop to the top and try to increment the
    // refcount of the existing object.

#if UPB_TSAN
    // Yield to give other threads a chance to increment the refcount.  This is
    // especially an issue for TSAN builds, which are prone to locking up from
    // the thread with the upb_Atomic_Store call above getting starved.
    sched_yield();
#endif  // UPB_TSAN
  }
}

void upb_GeneratedRegistry_Release(const upb_GeneratedRegistryRef* r) {
  if (r == NULL) return;

  upb_GeneratedRegistry* registry = _upb_generated_registry();

  int ref_count = upb_Atomic_Sub(&registry->ref_count, 1, memory_order_acq_rel);
  UPB_ASSERT(registry->ref_count >= 0);

  // A ref_count of 1 means that we decremented the refcount to 0.
  if (ref_count == 1) {
    upb_GeneratedRegistryRef* ref =
        upb_Atomic_Exchange(&registry->ref, NULL, memory_order_acq_rel);
    if (ref != NULL) {
      // This is the last reference and we won any potential race to store NULL,
      // so we need to clean up.
      upb_Arena_Free(ref->UPB_PRIVATE(arena));
      upb_gfree(ref);
    }
  }
}

const upb_ExtensionRegistry* upb_GeneratedRegistry_Get(
    const upb_GeneratedRegistryRef* r) {
  if (r == NULL) return NULL;
  return r->UPB_PRIVATE(registry);
}


#include <stddef.h>
#include <stdint.h>


// Must be last.

const upb_MiniTableField* upb_MiniTable_GetOneof(const upb_MiniTable* m,
                                                 const upb_MiniTableField* f) {
  if (UPB_UNLIKELY(!upb_MiniTableField_IsInOneof(f))) {
    return NULL;
  }
  const upb_MiniTableField* ptr = &m->UPB_PRIVATE(fields)[0];
  const upb_MiniTableField* end =
      &m->UPB_PRIVATE(fields)[m->UPB_PRIVATE(field_count)];
  for (; ptr < end; ptr++) {
    if (ptr->presence == (*f).presence) {
      return ptr;
    }
  }
  return NULL;
}

bool upb_MiniTable_NextOneofField(const upb_MiniTable* m,
                                  const upb_MiniTableField** f) {
  const upb_MiniTableField* ptr = *f;
  const upb_MiniTableField* end =
      &m->UPB_PRIVATE(fields)[m->UPB_PRIVATE(field_count)];
  while (++ptr < end) {
    if (ptr->presence == (*f)->presence) {
      *f = ptr;
      return true;
    }
  }
  return false;
}

#include <stdio.h>
#include <stdlib.h>

// Must be last.

UPB_NORETURN void _upb_UnreachableFailure(const char* file, int line,
                                          const char* function_name) {
  fprintf(stderr, "%s:%d: Reached unreachable statement in function `%s`.\n",
          file, line, function_name ? function_name : "(unknown)");
  abort();
}


#include <assert.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Our awkward dance for including fasttable only when it is enabled.
#if UPB_FASTTABLE
#define UPB_INCLUDE_FAST_DECODE
#endif

#ifdef UPB_INCLUDE_FAST_DECODE
#endif

#undef UPB_INCLUDE_FAST_DECODE

// Must be last.

// A few fake field types for our tables.
enum {
  kUpb_FakeFieldType_FieldNotFound = 0,
  kUpb_FakeFieldType_MessageSetItem = 19,
};

// DecodeOp: an action to be performed for a wire-type/field-type combination.
enum {
  // Special ops: we don't write data to regular fields for these.
  kUpb_DecodeOp_UnknownField = -1,
  kUpb_DecodeOp_MessageSetItem = -2,

  // Scalar-only ops.
  kUpb_DecodeOp_Scalar1Byte = 0,
  kUpb_DecodeOp_Scalar4Byte = 2,
  kUpb_DecodeOp_Scalar8Byte = 3,

  // Scalar/repeated ops.
  kUpb_DecodeOp_String = 4,
  kUpb_DecodeOp_Bytes = 5,
  kUpb_DecodeOp_SubMessage = 6,

  // Repeated-only ops (also see macros below).
  kUpb_DecodeOp_PackedEnum = 13,
};

// For packed fields it is helpful to be able to recover the lg2 of the data
// size from the op.
#define OP_FIXPCK_LG2(n) (n + 5) /* n in [2, 3] => op in [7, 8] */
#define OP_VARPCK_LG2(n) (n + 9) /* n in [0, 2, 3] => op in [9, 11, 12] */

static bool _upb_Decoder_Reserve(upb_Decoder* d, upb_Array* arr, size_t elem) {
  bool need_realloc =
      arr->UPB_PRIVATE(capacity) - arr->UPB_PRIVATE(size) < elem;
  if (need_realloc && !UPB_PRIVATE(_upb_Array_Realloc)(
                          arr, arr->UPB_PRIVATE(size) + elem, &d->arena)) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
  }
  return need_realloc;
}

typedef struct {
  const char* ptr;
  uint64_t val;
} _upb_DecodeLongVarintReturn;

// This is identical to _upb_Decoder_DecodeTag() except that the maximum value
// is INT32_MAX instead of UINT32_MAX.
UPB_FORCEINLINE
const char* upb_Decoder_DecodeSize(upb_Decoder* d, const char* ptr,
                                   uint32_t* size) {
  int sz;
  ptr = upb_WireReader_ReadSize(ptr, &sz, EPS(d));
  *size = sz;
  return ptr;
}

UPB_FORCEINLINE upb_AddUnknownMode
_upb_Decoder_GetAddUnknownMode(upb_Decoder* d, const char* data) {
  if (d->options & kUpb_DecodeOption_AliasString) {
    if (data != d->input.buffer_start) {
      // If the data is not from the beginning of the input buffer, then we can
      // safely attempt to coalesce this region with the previous one.
      return kUpb_AddUnknown_AliasAllowMerge;
    } else {
      return kUpb_AddUnknown_Alias;
    }
  }
  return kUpb_AddUnknown_Copy;
}

static void _upb_Decoder_MungeInt32(wireval* val) {
  if (!upb_IsLittleEndian()) {
    /* The next stage will memcpy(dst, &val, 4) */
    val->uint32_val = val->uint64_val;
  }
}

static void _upb_Decoder_Munge(const upb_MiniTableField* field, wireval* val) {
  switch (field->UPB_PRIVATE(descriptortype)) {
    case kUpb_FieldType_Bool:
      val->bool_val = val->uint64_val != 0;
      break;
    case kUpb_FieldType_SInt32: {
      uint32_t n = val->uint64_val;
      val->uint32_val = (n >> 1) ^ -(int32_t)(n & 1);
      break;
    }
    case kUpb_FieldType_SInt64: {
      uint64_t n = val->uint64_val;
      val->uint64_val = (n >> 1) ^ -(int64_t)(n & 1);
      break;
    }
    case kUpb_FieldType_Int32:
    case kUpb_FieldType_UInt32:
      _upb_Decoder_MungeInt32(val);
      break;
    case kUpb_FieldType_Enum:
      UPB_UNREACHABLE();
  }
}

static upb_Message* _upb_Decoder_NewSubMessage2(upb_Decoder* d,
                                                const upb_MiniTable* subl,
                                                const upb_MiniTableField* field,
                                                upb_Message** target) {
  UPB_ASSERT(subl);
  upb_Message* msg = _upb_Message_New(subl, &d->arena);
  if (!msg) upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);

  *target = msg;
  return msg;
}

static upb_Message* _upb_Decoder_NewSubMessage(upb_Decoder* d,
                                               const upb_MiniTableField* field,
                                               upb_Message** target) {
  const upb_MiniTable* subl = upb_MiniTable_GetSubMessageTable(field);
  return _upb_Decoder_NewSubMessage2(d, subl, field, target);
}

static const char* _upb_Decoder_ReadString2(upb_Decoder* d, const char* ptr,
                                            int size, upb_StringView* str,
                                            bool validate_utf8) {
  if (!_upb_Decoder_ReadString(d, &ptr, size, str, validate_utf8)) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
  }
  return ptr;
}

/* libpg_query local patch: also stop recursing (with a MaxDepthExceeded error)
 * when the stack depth limit is reached, using Postgres' stack_is_too_deep().
 * The fixed depth limit alone can't prevent stack overflows, since the stack
 * available and the stack used per nesting level vary. See vendor/upb/README. */
extern bool stack_is_too_deep(void);

UPB_FORCEINLINE
const char* _upb_Decoder_RecurseSubMessage(upb_Decoder* d, const char* ptr,
                                           upb_Message* submsg,
                                           const upb_MiniTable* subl,
                                           uint32_t expected_end_group) {
  if (--d->depth < 0 || stack_is_too_deep()) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_MaxDepthExceeded);
  }
  ptr = _upb_Decoder_DecodeMessage(d, ptr, submsg, subl);
  d->depth++;
  if (d->end_group != expected_end_group) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
  }
  return ptr;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeSubMessage(upb_Decoder* d, const char* ptr,
                                          upb_Message* submsg,
                                          const upb_MiniTableField* field,
                                          size_t size) {
  ptrdiff_t delta = upb_EpsCopyInputStream_PushLimit(&d->input, ptr, size);
  const upb_MiniTable* subl = upb_MiniTable_GetSubMessageTable(field);
  UPB_ASSERT(subl);
  ptr = _upb_Decoder_RecurseSubMessage(d, ptr, submsg, subl, DECODE_NOGROUP);
  upb_EpsCopyInputStream_PopLimit(&d->input, ptr, delta);
  return ptr;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeGroup(upb_Decoder* d, const char* ptr,
                                     upb_Message* submsg,
                                     const upb_MiniTable* subl,
                                     uint32_t number) {
  if (upb_EpsCopyInputStream_IsDone(EPS(d), &ptr)) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
  }
  ptr = _upb_Decoder_RecurseSubMessage(d, ptr, submsg, subl, number);
  d->end_group = DECODE_NOGROUP;
  return ptr;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeKnownGroup(upb_Decoder* d, const char* ptr,
                                          upb_Message* submsg,
                                          const upb_MiniTableField* field) {
  const upb_MiniTable* subl = upb_MiniTable_GetSubMessageTable(field);
  UPB_ASSERT(subl);
  return _upb_Decoder_DecodeGroup(d, ptr, submsg, subl,
                                  field->UPB_PRIVATE(number));
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeFixedPacked(upb_Decoder* d, const char* ptr,
                                           upb_Array* arr, wireval* val,
                                           const upb_MiniTableField* field,
                                           int lg2) {
  upb_StringView sv;
  ptr = upb_EpsCopyInputStream_ReadStringEphemeral(&d->input, ptr, val->size,
                                                   &sv);
  if (!ptr) upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
  int mask = (1 << lg2) - 1;
  if (UPB_UNLIKELY((val->size & mask) != 0 || ptr == NULL)) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
  }
  size_t count = val->size >> lg2;
  if (count == 0) return ptr;
  _upb_Decoder_Reserve(d, arr, count);
  void* mem = UPB_PTR_AT(upb_Array_MutableDataPtr(arr),
                         arr->UPB_PRIVATE(size) << lg2, void);
  arr->UPB_PRIVATE(size) += count;
  if (upb_IsLittleEndian()) {
    memcpy(mem, sv.data, sv.size);
  } else {
    const char* src = sv.data;
    const char* src_end = src + sv.size;
    char* dst = mem;
    if (lg2 == 2) {
      for (; src < src_end; src += 4, dst += 4) {
        uint32_t x;
        memcpy(&x, src, 4);
        x = upb_BigEndian32(x);
        memcpy(dst, &x, 4);
      }
    } else {
      UPB_ASSERT(lg2 == 3);
      for (; src < src_end; src += 8, dst += 8) {
        uint64_t x;
        memcpy(&x, src, 8);
        x = upb_BigEndian64(x);
        memcpy(dst, &x, 8);
      }
    }
  }

  return ptr;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeVarintPacked(upb_Decoder* d, const char* ptr,
                                            upb_Array* arr, wireval* val,
                                            const upb_MiniTableField* field,
                                            int lg2) {
  int scale = 1 << lg2;
  ptrdiff_t delta = upb_EpsCopyInputStream_PushLimit(&d->input, ptr, val->size);
  char* out = UPB_PTR_AT(upb_Array_MutableDataPtr(arr),
                         arr->UPB_PRIVATE(size) << lg2, void);
  while (!upb_EpsCopyInputStream_IsDone(EPS(d), &ptr)) {
    wireval elem;
    ptr = upb_WireReader_ReadVarint(ptr, &elem.uint64_val, EPS(d));
    _upb_Decoder_Munge(field, &elem);
    if (_upb_Decoder_Reserve(d, arr, 1)) {
      out = UPB_PTR_AT(upb_Array_MutableDataPtr(arr),
                       arr->UPB_PRIVATE(size) << lg2, void);
    }
    arr->UPB_PRIVATE(size)++;
    memcpy(out, &elem, scale);
    out += scale;
  }
  upb_EpsCopyInputStream_PopLimit(&d->input, ptr, delta);
  return ptr;
}

UPB_NOINLINE
static const char* _upb_Decoder_DecodeEnumPacked(
    upb_Decoder* d, const char* ptr, upb_Message* msg, upb_Array* arr,
    const upb_MiniTableField* field, wireval* val) {
  const upb_MiniTableEnum* e = upb_MiniTable_GetSubEnumTable(field);
  ptrdiff_t delta = upb_EpsCopyInputStream_PushLimit(&d->input, ptr, val->size);
  char* out = UPB_PTR_AT(upb_Array_MutableDataPtr(arr),
                         arr->UPB_PRIVATE(size) * 4, void);
  while (!upb_EpsCopyInputStream_IsDone(EPS(d), &ptr)) {
    wireval elem;
    ptr = upb_WireReader_ReadVarint(ptr, &elem.uint64_val, EPS(d));
    if (!upb_MiniTableEnum_CheckValue(e, elem.uint64_val)) {
      upb_Message* unknown_msg =
          field->UPB_PRIVATE(mode) & kUpb_LabelFlags_IsExtension
              ? d->original_msg
              : msg;
      if (!_upb_Encoder_AddEnumValueToUnknown(unknown_msg, field,
                                              elem.uint64_val, &d->arena)) {
        upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
      }
      continue;
    }
    _upb_Decoder_MungeInt32(&elem);
    if (_upb_Decoder_Reserve(d, arr, 1)) {
      out = UPB_PTR_AT(upb_Array_MutableDataPtr(arr),
                       arr->UPB_PRIVATE(size) * 4, void);
    }
    arr->UPB_PRIVATE(size)++;
    memcpy(out, &elem, 4);
    out += 4;
  }
  upb_EpsCopyInputStream_PopLimit(&d->input, ptr, delta);
  return ptr;
}

static upb_Array* _upb_Decoder_CreateArray(upb_Decoder* d,
                                           const upb_MiniTableField* field) {
  const upb_FieldType field_type = field->UPB_PRIVATE(descriptortype);
  const size_t lg2 = UPB_PRIVATE(_upb_FieldType_SizeLg2)(field_type);
  upb_Array* ret = UPB_PRIVATE(_upb_Array_New)(
      &d->arena, _UPB_ARRAY_DEFAULT_INITIAL_SIZE, lg2);
  if (!ret) upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
  return ret;
}

#if UPB_FASTTABLE
UPB_PRESERVE_NONE
#endif
static const char* _upb_Decoder_DecodeToArray(upb_Decoder* d, const char* ptr,
                                              upb_Message* msg,
                                              const upb_MiniTableField* field,
                                              wireval* val, int op) {
  upb_Array** arrp = UPB_PTR_AT(msg, field->UPB_PRIVATE(offset), void);
  upb_Array* arr = *arrp;
  void* mem;

  if (arr) {
    _upb_Decoder_Reserve(d, arr, 1);
  } else {
    arr = _upb_Decoder_CreateArray(d, field);
    *arrp = arr;
  }

  switch (op) {
    case kUpb_DecodeOp_Scalar1Byte:
    case kUpb_DecodeOp_Scalar4Byte:
    case kUpb_DecodeOp_Scalar8Byte:
      /* Append scalar value. */
      mem = UPB_PTR_AT(upb_Array_MutableDataPtr(arr),
                       arr->UPB_PRIVATE(size) << op, void);
      arr->UPB_PRIVATE(size)++;
      memcpy(mem, val, 1 << op);
      return ptr;
    case kUpb_DecodeOp_String: {
      /* Append string. */
      upb_StringView* str = (upb_StringView*)upb_Array_MutableDataPtr(arr) +
                            arr->UPB_PRIVATE(size);
      ptr = _upb_Decoder_ReadString2(d, ptr, val->size, str,
                                     /*validate_utf8=*/true);
      arr->UPB_PRIVATE(size)++;
      return ptr;
    }
    case kUpb_DecodeOp_Bytes: {
      /* Append bytes. */
      upb_StringView* str = (upb_StringView*)upb_Array_MutableDataPtr(arr) +
                            arr->UPB_PRIVATE(size);
      ptr = _upb_Decoder_ReadString2(d, ptr, val->size, str,
                                     /*validate_utf8=*/false);
      arr->UPB_PRIVATE(size)++;
      return ptr;
    }
    case kUpb_DecodeOp_SubMessage: {
      /* Append submessage / group. */
      upb_Message** target =
          UPB_PTR_AT(upb_Array_MutableDataPtr(arr),
                     arr->UPB_PRIVATE(size) * sizeof(void*), upb_Message*);
      upb_Message* submsg = _upb_Decoder_NewSubMessage(d, field, target);
      arr->UPB_PRIVATE(size)++;
      if (UPB_UNLIKELY(field->UPB_PRIVATE(descriptortype) ==
                       kUpb_FieldType_Group)) {
        return _upb_Decoder_DecodeKnownGroup(d, ptr, submsg, field);
      } else {
        return _upb_Decoder_DecodeSubMessage(d, ptr, submsg, field, val->size);
      }
    }
    case OP_FIXPCK_LG2(2):
    case OP_FIXPCK_LG2(3):
      return _upb_Decoder_DecodeFixedPacked(d, ptr, arr, val, field,
                                            op - OP_FIXPCK_LG2(0));
    case OP_VARPCK_LG2(0):
    case OP_VARPCK_LG2(2):
    case OP_VARPCK_LG2(3):
      return _upb_Decoder_DecodeVarintPacked(d, ptr, arr, val, field,
                                             op - OP_VARPCK_LG2(0));
    case kUpb_DecodeOp_PackedEnum:
      return _upb_Decoder_DecodeEnumPacked(d, ptr, msg, arr, field, val);
    default:
      UPB_UNREACHABLE();
  }
}

static upb_Map* _upb_Decoder_CreateMap(upb_Decoder* d,
                                       const upb_MiniTable* entry) {
  // Maps descriptor type -> upb map size
  static const uint8_t kSizeInMap[] = {
      [0] = -1,  // invalid descriptor type
      [kUpb_FieldType_Double] = 8,
      [kUpb_FieldType_Float] = 4,
      [kUpb_FieldType_Int64] = 8,
      [kUpb_FieldType_UInt64] = 8,
      [kUpb_FieldType_Int32] = 4,
      [kUpb_FieldType_Fixed64] = 8,
      [kUpb_FieldType_Fixed32] = 4,
      [kUpb_FieldType_Bool] = 1,
      [kUpb_FieldType_String] = UPB_MAPTYPE_STRING,
      [kUpb_FieldType_Group] = sizeof(void*),
      [kUpb_FieldType_Message] = sizeof(void*),
      [kUpb_FieldType_Bytes] = UPB_MAPTYPE_STRING,
      [kUpb_FieldType_UInt32] = 4,
      [kUpb_FieldType_Enum] = 4,
      [kUpb_FieldType_SFixed32] = 4,
      [kUpb_FieldType_SFixed64] = 8,
      [kUpb_FieldType_SInt32] = 4,
      [kUpb_FieldType_SInt64] = 8,
  };

  const upb_MiniTableField* key_field = &entry->UPB_PRIVATE(fields)[0];
  const upb_MiniTableField* val_field = &entry->UPB_PRIVATE(fields)[1];
  char key_size = kSizeInMap[key_field->UPB_PRIVATE(descriptortype)];
  char val_size = kSizeInMap[val_field->UPB_PRIVATE(descriptortype)];
  UPB_ASSERT(key_field->UPB_PRIVATE(offset) == offsetof(upb_MapEntry, k));
  UPB_ASSERT(val_field->UPB_PRIVATE(offset) == offsetof(upb_MapEntry, v));
  upb_Map* ret = _upb_Map_New(&d->arena, key_size, val_size);
  if (!ret) upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
  return ret;
}

#if UPB_FASTTABLE
UPB_PRESERVE_NONE
#endif
static const char* _upb_Decoder_DecodeToMap(upb_Decoder* d, const char* ptr,
                                            upb_Message* msg,
                                            const upb_MiniTableField* field,
                                            wireval* val) {
  upb_Map** map_p = UPB_PTR_AT(msg, field->UPB_PRIVATE(offset), upb_Map*);
  upb_Map* map = *map_p;
  upb_MapEntry ent;
  UPB_ASSERT(upb_MiniTableField_Type(field) == kUpb_FieldType_Message);
  const upb_MiniTable* entry = upb_MiniTable_GetSubMessageTable(field);

  UPB_ASSERT(entry);
  UPB_ASSERT(entry->UPB_PRIVATE(field_count) == 2);
  UPB_ASSERT(upb_MiniTableField_IsScalar(&entry->UPB_PRIVATE(fields)[0]));
  UPB_ASSERT(upb_MiniTableField_IsScalar(&entry->UPB_PRIVATE(fields)[1]));

  if (!map) {
    map = _upb_Decoder_CreateMap(d, entry);
    *map_p = map;
  }

  // Parse map entry.
  memset(&ent, 0, sizeof(ent));

  bool value_is_message =
      entry->UPB_PRIVATE(fields)[1].UPB_PRIVATE(descriptortype) ==
          kUpb_FieldType_Message ||
      entry->UPB_PRIVATE(fields)[1].UPB_PRIVATE(descriptortype) ==
          kUpb_FieldType_Group;
  const upb_MiniTable* sub_table =
      value_is_message
          ? upb_MiniTable_GetSubMessageTable(&entry->UPB_PRIVATE(fields)[1])
          : NULL;
  upb_Message* sub_msg = NULL;

  if (sub_table) {
    // Create proactively to handle the case where it doesn't appear.
    _upb_Decoder_NewSubMessage(d, &entry->UPB_PRIVATE(fields)[1], &sub_msg);
    ent.v.val = upb_value_ptr(sub_msg);
  }

  ptr = _upb_Decoder_DecodeSubMessage(d, ptr, &ent.message, field, val->size);

  if (sub_msg && sub_table->UPB_PRIVATE(required_count)) {
    // If the map entry did not contain a value on the wire, `sub_msg` is an
    // empty message; we must check if it is missing any required fields. If the
    // value was present, this check is redundant but harmless.
    _upb_Decoder_CheckRequired(d, ptr, sub_msg, sub_table);
  }

  if (upb_Message_HasUnknown(&ent.message)) {
    if (!_upb_Encoder_AddMapEntryUnknown(msg, field, &ent.message, entry,
                                         &d->arena)) {
      upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
    }
  } else {
    if (_upb_Map_Insert(map, &ent.k, map->key_size, &ent.v, map->val_size,
                        &d->arena) == kUpb_MapInsertStatus_OutOfMemory) {
      upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
    }
  }
  return ptr;
}

#if UPB_FASTTABLE
UPB_PRESERVE_NONE
#endif
static const char* _upb_Decoder_DecodeToSubMessage(
    upb_Decoder* d, const char* ptr, upb_Message* msg,
    const upb_MiniTableField* field, wireval* val, int op) {
  void* mem = UPB_PTR_AT(msg, field->UPB_PRIVATE(offset), void);
  int type = field->UPB_PRIVATE(descriptortype);

  // Set presence if necessary.
  if (UPB_PRIVATE(_upb_MiniTableField_HasHasbit)(field)) {
    UPB_PRIVATE(_upb_Message_SetHasbit)(msg, field);
  } else if (upb_MiniTableField_IsInOneof(field)) {
    // Oneof case
    uint32_t* oneof_case = UPB_PRIVATE(_upb_Message_OneofCasePtr)(msg, field);
    if (op == kUpb_DecodeOp_SubMessage &&
        *oneof_case != field->UPB_PRIVATE(number)) {
      memset(mem, 0, sizeof(void*));
    }
    *oneof_case = field->UPB_PRIVATE(number);
  }

  // Store into message.
  switch (op) {
    case kUpb_DecodeOp_SubMessage: {
      upb_Message** submsgp = mem;
      upb_Message* submsg = *submsgp;
      if (!submsg) submsg = _upb_Decoder_NewSubMessage(d, field, submsgp);
      if (UPB_UNLIKELY(type == kUpb_FieldType_Group)) {
        ptr = _upb_Decoder_DecodeKnownGroup(d, ptr, submsg, field);
      } else {
        ptr = _upb_Decoder_DecodeSubMessage(d, ptr, submsg, field, val->size);
      }
      break;
    }
    case kUpb_DecodeOp_String:
      return _upb_Decoder_ReadString2(d, ptr, val->size, mem,
                                      /*validate_utf8=*/true);
    case kUpb_DecodeOp_Bytes:
      return _upb_Decoder_ReadString2(d, ptr, val->size, mem,
                                      /*validate_utf8=*/false);
    case kUpb_DecodeOp_Scalar8Byte:
      memcpy(mem, val, 8);
      break;
    case kUpb_DecodeOp_Scalar4Byte:
      memcpy(mem, val, 4);
      break;
    case kUpb_DecodeOp_Scalar1Byte:
      memcpy(mem, val, 1);
      break;
    default:
      UPB_UNREACHABLE();
  }

  return ptr;
}

enum {
  kStartItemTag = ((kUpb_MsgSet_Item << 3) | kUpb_WireType_StartGroup),
  kEndItemTag = ((kUpb_MsgSet_Item << 3) | kUpb_WireType_EndGroup),
  kTypeIdTag = ((kUpb_MsgSet_TypeId << 3) | kUpb_WireType_Varint),
  kMessageTag = ((kUpb_MsgSet_Message << 3) | kUpb_WireType_Delimited),
};

static void upb_Decoder_AddKnownMessageSetItem(
    upb_Decoder* d, upb_Message* msg, const upb_MiniTableExtension* item_mt,
    const char* data, uint32_t size) {
  upb_Extension* ext =
      UPB_PRIVATE(_upb_Message_GetOrCreateExtension)(msg, item_mt, &d->arena);
  if (UPB_UNLIKELY(!ext)) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
  }
  upb_Message** submsgp = (upb_Message**)&ext->data.msg_val;
  upb_Message* submsg = _upb_Decoder_NewSubMessage2(
      d, ext->ext->UPB_PRIVATE(sub).UPB_PRIVATE(submsg),
      &ext->ext->UPB_PRIVATE(field), submsgp);
  // upb_Decode_LimitDepth() takes uint32_t, d->depth - 1 can not be negative.
  if (d->depth <= 1) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_MaxDepthExceeded);
  }
  upb_DecodeStatus status = upb_Decode(
      data, size, submsg, upb_MiniTableExtension_GetSubMessage(item_mt),
      d->extreg, upb_Decode_LimitDepth(d->options, d->depth - 1), &d->arena);
  if (status != kUpb_DecodeStatus_Ok) {
    upb_ErrorHandler_ThrowError(d->err, status);
  }
}

static void upb_Decoder_AddUnknownMessageSetItem(upb_Decoder* d,
                                                 upb_Message* msg,
                                                 uint32_t type_id,
                                                 const char* message_data,
                                                 uint32_t message_size) {
  char buf[6 * kUpb_Encoder_EncodeVarint32MaxSize];
  char* ptr = buf;
  ptr = upb_Encoder_EncodeVarint32(kStartItemTag, ptr);
  ptr = upb_Encoder_EncodeVarint32(kTypeIdTag, ptr);
  ptr = upb_Encoder_EncodeVarint32(type_id, ptr);
  ptr = upb_Encoder_EncodeVarint32(kMessageTag, ptr);
  ptr = upb_Encoder_EncodeVarint32(message_size, ptr);
  char* split = ptr;

  ptr = upb_Encoder_EncodeVarint32(kEndItemTag, ptr);
  char* end = ptr;
  upb_StringView unknown[] = {
      {buf, split - buf},
      {message_data, message_size},
      {split, end - split},
  };
  if (!UPB_PRIVATE(_upb_Message_AddUnknownV)(msg, &d->arena, unknown, 3)) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
  }
}

static void upb_Decoder_AddMessageSetItem(upb_Decoder* d, upb_Message* msg,
                                          const upb_MiniTable* t,
                                          uint32_t type_id, const char* data,
                                          uint32_t size) {
  const upb_MiniTableExtension* item_mt =
      upb_ExtensionRegistry_Lookup(d->extreg, t, type_id);
  if (item_mt) {
    upb_Decoder_AddKnownMessageSetItem(d, msg, item_mt, data, size);
  } else {
    upb_Decoder_AddUnknownMessageSetItem(d, msg, type_id, data, size);
  }
}

static const char* upb_Decoder_DecodeMessageSetItem(
    upb_Decoder* d, const char* ptr, upb_Message* msg,
    const upb_MiniTable* layout) {
  uint32_t type_id = 0;
  upb_StringView preserved = {NULL, 0};
  typedef enum {
    kUpb_HaveId = 1 << 0,
    kUpb_HavePayload = 1 << 1,
  } StateMask;
  StateMask state_mask = 0;
  while (!upb_EpsCopyInputStream_IsDone(EPS(d), &ptr)) {
    uint32_t tag;
    ptr = upb_WireReader_ReadTag(ptr, &tag, EPS(d));
    switch (tag) {
      case kEndItemTag:
        return ptr;
      case kTypeIdTag: {
        uint64_t tmp;
        ptr = upb_WireReader_ReadVarint(ptr, &tmp, EPS(d));
        if (state_mask & kUpb_HaveId) break;  // Ignore dup.
        state_mask |= kUpb_HaveId;
        type_id = tmp;
        if (state_mask & kUpb_HavePayload) {
          upb_Decoder_AddMessageSetItem(d, msg, layout, type_id, preserved.data,
                                        preserved.size);
        }
        break;
      }
      case kMessageTag: {
        uint32_t size;
        upb_StringView sv;
        ptr = upb_Decoder_DecodeSize(d, ptr, &size);
        ptr = upb_EpsCopyInputStream_ReadStringAlwaysAlias(&d->input, ptr, size,
                                                           &sv);
        if (!ptr) {
          upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
        }
        if (state_mask & kUpb_HavePayload) break;  // Ignore dup.
        state_mask |= kUpb_HavePayload;
        if (state_mask & kUpb_HaveId) {
          upb_Decoder_AddMessageSetItem(d, msg, layout, type_id, sv.data,
                                        sv.size);
        } else {
          // Out of order, we must preserve the payload.
          preserved = sv;
        }
        break;
      }
      default:
        // We do not preserve unexpected fields inside a message set item.
        ptr = _upb_WireReader_SkipValue(ptr, tag, d->depth, &d->input);
        break;
    }
  }
  upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
}

static upb_MiniTableField upb_Decoder_FieldNotFoundField = {
    0, 0, 0, 0, kUpb_FakeFieldType_FieldNotFound, 0};

UPB_NOINLINE const upb_MiniTableField* _upb_Decoder_FindExtensionField(
    upb_Decoder* d, const upb_MiniTable* t, uint32_t field_number, int ext_mode,
    uint32_t wire_type) {
  // Treat a message set as an extendable message if it is a delimited field.
  // This provides compatibility with encoders that are unaware of message
  // sets and serialize them as normal extensions.
  if (ext_mode == kUpb_ExtMode_Extendable ||
      (ext_mode == kUpb_ExtMode_IsMessageSet &&
       wire_type == kUpb_WireType_Delimited)) {
    const upb_MiniTableExtension* ext =
        upb_ExtensionRegistry_Lookup(d->extreg, t, field_number);
    if (ext) return &ext->UPB_PRIVATE(field);
  } else if (ext_mode == kUpb_ExtMode_IsMessageSet) {
    if (field_number == kUpb_MsgSet_Item) {
      static upb_MiniTableField item = {
          0, 0, 0, 0, kUpb_FakeFieldType_MessageSetItem, 0};
      return &item;
    }
  }
  return &upb_Decoder_FieldNotFoundField;
}

static const upb_MiniTableField* _upb_Decoder_FindField(upb_Decoder* d,
                                                        const upb_MiniTable* t,
                                                        uint32_t field_number,
                                                        uint32_t wire_type) {
  UPB_ASSERT(t);
  const upb_MiniTableField* field =
      upb_MiniTable_FindFieldByNumber(t, field_number);
  if (field) return field;

  if (d->extreg && UPB_PRIVATE(_upb_MiniTable_ExtModeBase)(t)) {
    return _upb_Decoder_FindExtensionField(
        d, t, field_number, UPB_PRIVATE(_upb_MiniTable_ExtModeBase)(t),
        wire_type);
  }

  return &upb_Decoder_FieldNotFoundField;  // Unknown field.
}

static int _upb_Decoder_GetVarintOp(const upb_MiniTableField* field) {
  static const int8_t kVarintOps[] = {
      [kUpb_FakeFieldType_FieldNotFound] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Double] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Float] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Int64] = kUpb_DecodeOp_Scalar8Byte,
      [kUpb_FieldType_UInt64] = kUpb_DecodeOp_Scalar8Byte,
      [kUpb_FieldType_Int32] = kUpb_DecodeOp_Scalar4Byte,
      [kUpb_FieldType_Fixed64] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Fixed32] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Bool] = kUpb_DecodeOp_Scalar1Byte,
      [kUpb_FieldType_String] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Group] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Message] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Bytes] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_UInt32] = kUpb_DecodeOp_Scalar4Byte,
      [kUpb_FieldType_Enum] = kUpb_DecodeOp_Scalar4Byte,
      [kUpb_FieldType_SFixed32] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_SFixed64] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_SInt32] = kUpb_DecodeOp_Scalar4Byte,
      [kUpb_FieldType_SInt64] = kUpb_DecodeOp_Scalar8Byte,
      [kUpb_FakeFieldType_MessageSetItem] = kUpb_DecodeOp_UnknownField,
  };

  return kVarintOps[field->UPB_PRIVATE(descriptortype)];
}

UPB_FORCEINLINE
void _upb_Decoder_CheckUnlinked(upb_Decoder* d, const upb_MiniTable* mt,
                                const upb_MiniTableField* field, int* op) {
  // If sub-message is not linked, treat as unknown.
  if (field->UPB_PRIVATE(mode) & kUpb_LabelFlags_IsExtension) return;
  const upb_MiniTable* mt_sub = upb_MiniTable_GetSubMessageTable(field);
  if (mt_sub != NULL) return;  // Normal case, sub-message is linked.
  _upb_Decoder_VerifyOneofUnlinked(mt, field);
  *op = kUpb_DecodeOp_UnknownField;
}

UPB_FORCEINLINE
void _upb_Decoder_MaybeVerifyUtf8(upb_Decoder* d,
                                  const upb_MiniTableField* field, int* op) {
  UPB_ASSUME(field->UPB_PRIVATE(descriptortype) == kUpb_FieldType_Bytes);
  if (_upb_Decoder_FieldRequiresUtf8Validation(d, field)) {
    *op = kUpb_DecodeOp_String;
  }
}

static int _upb_Decoder_GetDelimitedOp(upb_Decoder* d, const upb_MiniTable* mt,
                                       const upb_MiniTableField* field) {
  enum { kRepeatedBase = 19 };

  static const int8_t kDelimitedOps[] = {
      // For non-repeated field type.
      [kUpb_FakeFieldType_FieldNotFound] =
          kUpb_DecodeOp_UnknownField,  // Field not found.
      [kUpb_FieldType_Double] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Float] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Int64] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_UInt64] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Int32] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Fixed64] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Fixed32] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Bool] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_String] = kUpb_DecodeOp_String,
      [kUpb_FieldType_Group] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Message] = kUpb_DecodeOp_SubMessage,
      [kUpb_FieldType_Bytes] = kUpb_DecodeOp_Bytes,
      [kUpb_FieldType_UInt32] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_Enum] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_SFixed32] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_SFixed64] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_SInt32] = kUpb_DecodeOp_UnknownField,
      [kUpb_FieldType_SInt64] = kUpb_DecodeOp_UnknownField,
      [kUpb_FakeFieldType_MessageSetItem] = kUpb_DecodeOp_SubMessage,
      // For repeated field type.
      [kRepeatedBase + kUpb_FieldType_Double] = OP_FIXPCK_LG2(3),
      [kRepeatedBase + kUpb_FieldType_Float] = OP_FIXPCK_LG2(2),
      [kRepeatedBase + kUpb_FieldType_Int64] = OP_VARPCK_LG2(3),
      [kRepeatedBase + kUpb_FieldType_UInt64] = OP_VARPCK_LG2(3),
      [kRepeatedBase + kUpb_FieldType_Int32] = OP_VARPCK_LG2(2),
      [kRepeatedBase + kUpb_FieldType_Fixed64] = OP_FIXPCK_LG2(3),
      [kRepeatedBase + kUpb_FieldType_Fixed32] = OP_FIXPCK_LG2(2),
      [kRepeatedBase + kUpb_FieldType_Bool] = OP_VARPCK_LG2(0),
      [kRepeatedBase + kUpb_FieldType_String] = kUpb_DecodeOp_String,
      [kRepeatedBase + kUpb_FieldType_Group] = kUpb_DecodeOp_UnknownField,
      [kRepeatedBase + kUpb_FieldType_Message] = kUpb_DecodeOp_SubMessage,
      [kRepeatedBase + kUpb_FieldType_Bytes] = kUpb_DecodeOp_Bytes,
      [kRepeatedBase + kUpb_FieldType_UInt32] = OP_VARPCK_LG2(2),
      [kRepeatedBase + kUpb_FieldType_Enum] = kUpb_DecodeOp_PackedEnum,
      [kRepeatedBase + kUpb_FieldType_SFixed32] = OP_FIXPCK_LG2(2),
      [kRepeatedBase + kUpb_FieldType_SFixed64] = OP_FIXPCK_LG2(3),
      [kRepeatedBase + kUpb_FieldType_SInt32] = OP_VARPCK_LG2(2),
      [kRepeatedBase + kUpb_FieldType_SInt64] = OP_VARPCK_LG2(3),
      // Omitting kUpb_FakeFieldType_MessageSetItem, because we never emit a
      // repeated msgset type
  };

  int ndx = field->UPB_PRIVATE(descriptortype);
  if (upb_MiniTableField_IsArray(field)) ndx += kRepeatedBase;
  int op = kDelimitedOps[ndx];

  if (op == kUpb_DecodeOp_SubMessage) {
    _upb_Decoder_CheckUnlinked(d, mt, field, &op);
  } else if (op == kUpb_DecodeOp_Bytes) {
    _upb_Decoder_MaybeVerifyUtf8(d, field, &op);
  }

  /* libpg_query local patch: decode proto3 string fields as bytes to skip
   * UTF-8 validation (SQL literals may be non-UTF-8). See vendor/upb/README. */
  if (op == kUpb_DecodeOp_String) op = kUpb_DecodeOp_Bytes;

  return op;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeWireValue(upb_Decoder* d, const char* ptr,
                                         const upb_MiniTable* mt,
                                         const upb_MiniTableField* field,
                                         uint32_t wire_type, wireval* val,
                                         int* op) {
  static const unsigned kFixed32OkMask = (1 << kUpb_FieldType_Float) |
                                         (1 << kUpb_FieldType_Fixed32) |
                                         (1 << kUpb_FieldType_SFixed32);

  static const unsigned kFixed64OkMask = (1 << kUpb_FieldType_Double) |
                                         (1 << kUpb_FieldType_Fixed64) |
                                         (1 << kUpb_FieldType_SFixed64);

  switch (wire_type) {
    case kUpb_WireType_Varint:
      ptr = upb_WireReader_ReadVarint(ptr, &val->uint64_val, EPS(d));
      if (upb_MiniTableField_IsClosedEnum(field)) {
        const upb_MiniTableEnum* e = upb_MiniTable_GetSubEnumTable(field);
        if (!upb_MiniTableEnum_CheckValue(e, val->uint64_val)) {
          *op = kUpb_DecodeOp_UnknownField;
          return ptr;
        }
        _upb_Decoder_MungeInt32(val);
      } else {
        _upb_Decoder_Munge(field, val);
      }
      *op = _upb_Decoder_GetVarintOp(field);
      return ptr;
    case kUpb_WireType_32Bit:
      *op = kUpb_DecodeOp_Scalar4Byte;
      if (((1 << field->UPB_PRIVATE(descriptortype)) & kFixed32OkMask) == 0) {
        *op = kUpb_DecodeOp_UnknownField;
      }
      return upb_WireReader_ReadFixed32(ptr, &val->uint32_val, &d->input);
    case kUpb_WireType_64Bit:
      *op = kUpb_DecodeOp_Scalar8Byte;
      if (((1 << field->UPB_PRIVATE(descriptortype)) & kFixed64OkMask) == 0) {
        *op = kUpb_DecodeOp_UnknownField;
      }
      return upb_WireReader_ReadFixed64(ptr, &val->uint64_val, &d->input);
    case kUpb_WireType_Delimited:
      ptr = upb_Decoder_DecodeSize(d, ptr, &val->size);
      *op = _upb_Decoder_GetDelimitedOp(d, mt, field);
      return ptr;
    case kUpb_WireType_StartGroup:
      val->uint32_val = field->UPB_PRIVATE(number);
      if (field->UPB_PRIVATE(descriptortype) == kUpb_FieldType_Group) {
        *op = kUpb_DecodeOp_SubMessage;
        _upb_Decoder_CheckUnlinked(d, mt, field, op);
      } else if (field->UPB_PRIVATE(descriptortype) ==
                 kUpb_FakeFieldType_MessageSetItem) {
        *op = kUpb_DecodeOp_MessageSetItem;
      } else {
        *op = kUpb_DecodeOp_UnknownField;
      }
      return ptr;
    default:
      break;
  }
  upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeKnownField(upb_Decoder* d, const char* ptr,
                                          upb_Message* msg,
                                          const upb_MiniTableField* field,
                                          int op, wireval* val) {
  uint8_t mode = field->UPB_PRIVATE(mode);

  if (UPB_UNLIKELY(mode & kUpb_LabelFlags_IsExtension)) {
    const upb_MiniTableExtension* ext_layout =
        (const upb_MiniTableExtension*)field;
    upb_Extension* ext = UPB_PRIVATE(_upb_Message_GetOrCreateExtension)(
        msg, ext_layout, &d->arena);
    if (UPB_UNLIKELY(!ext)) {
      upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
    }
    d->original_msg = msg;
    msg = &ext->data.UPB_PRIVATE(ext_msg_val);
  }

  switch (mode & kUpb_FieldMode_Mask) {
    case kUpb_FieldMode_Array:
      return _upb_Decoder_DecodeToArray(d, ptr, msg, field, val, op);
    case kUpb_FieldMode_Map:
      return _upb_Decoder_DecodeToMap(d, ptr, msg, field, val);
    case kUpb_FieldMode_Scalar:
      return _upb_Decoder_DecodeToSubMessage(d, ptr, msg, field, val, op);
    default:
      UPB_UNREACHABLE();
  }
}

UPB_FORCEINLINE
bool _upb_Decoder_CanSkipUnknownField(upb_Decoder* d, uint32_t next_field_num,
                                      uint32_t next_wire_type,
                                      const upb_MiniTable* mt, uint32_t* gap_lo,
                                      uint32_t* gap_hi, bool is_extendable) {
  if (next_wire_type == kUpb_WireType_EndGroup) {
    return false;
  }

  if (UPB_UNLIKELY(upb_MiniTable_IsMessageSet(mt))) {
    if (next_field_num == kUpb_MsgSet_Item) {
      return false;
    }
  }

  if (next_field_num <= *gap_lo || next_field_num >= *gap_hi) {
    if (UPB_LIKELY(next_field_num == *gap_hi)) {
      // Common case of fields in ascending order encountering a known
      // field
      return false;
    }
    if (UPB_UNLIKELY(next_field_num == 0)) {
      upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
    }
    if (next_field_num == *gap_lo) {
      return false;
    }
    if (!UPB_PRIVATE(_upb_MiniTable_FindUnknownGap)(mt, next_field_num, gap_lo,
                                                    gap_hi)) {
      return false;
    }
  }

  if (is_extendable) {
    if (upb_ExtensionRegistry_Lookup(d->extreg, mt, next_field_num)) {
      return false;
    }
  }

  return true;
}

UPB_NOINLINE
static const char* _upb_Decoder_DecodeUnknowns(
    upb_Decoder* d, const char* ptr, upb_Message* msg, const upb_MiniTable* mt,
    uint32_t field_number, uint32_t wire_type, wireval val, const char* start) {
  if (field_number == 0) {
    upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_Malformed);
  }

  upb_EpsCopyCapture capture;
  upb_EpsCopyCapture_Start(&capture, &d->input, start);

  // We have already parsed the tag and "value" (or size) of the first unknown
  // field. ptr is currently:
  // - after value for Varint, 32Bit, 64Bit.
  // - after size for Delimited.
  // - after tag for StartGroup.

  // We need to finish skipping the first field.
  if (wire_type == kUpb_WireType_Delimited) {
    upb_StringView sv;
    ptr = upb_EpsCopyInputStream_ReadStringEphemeral(&d->input, ptr, val.size,
                                                     &sv);
  } else if (wire_type == kUpb_WireType_StartGroup) {
    ptr = UPB_PRIVATE(_upb_WireReader_SkipGroup)(ptr, field_number << 3,
                                                 d->depth, &d->input);
  }

  // Fast check if next fields are also unknown, based on the gap between known
  // fields our first unknown falls between.

  uint32_t gap_lo = 0;
  uint32_t gap_hi = 0;
  bool has_gap = UPB_PRIVATE(_upb_MiniTable_FindUnknownGap)(mt, field_number,
                                                            &gap_lo, &gap_hi);

  if (has_gap) {
    bool is_extendable =
        (UPB_UNLIKELY(UPB_PRIVATE(_upb_MiniTable_IsExtendable)(mt)) ||
         UPB_UNLIKELY(upb_MiniTable_IsMessageSet(mt))) &&
        d->extreg;
    while (!upb_EpsCopyInputStream_IsDone(EPS(d), &ptr)) {
      const char* start_ptr = ptr;
      uint32_t tag;
      ptr = upb_WireReader_ReadTag(ptr, &tag, EPS(d));

      uint32_t next_field_num = tag >> 3;
      uint32_t next_wire_type = tag & 7;

      if (_upb_Decoder_CanSkipUnknownField(d, next_field_num, next_wire_type,
                                           mt, &gap_lo, &gap_hi,
                                           is_extendable)) {
        ptr = _upb_WireReader_SkipValueForceInline(ptr, tag, d->depth, EPS(d));
      } else {
        ptr = start_ptr;
        break;
      }
    }
  }

  upb_StringView sv;
  upb_EpsCopyCapture_End(&capture, &d->input, ptr, &sv);

  if (UPB_LIKELY(sv.size > 0)) {
    if (!UPB_PRIVATE(_upb_Message_AddUnknown)(
            msg, sv.data, sv.size, &d->arena,
            _upb_Decoder_GetAddUnknownMode(d, sv.data))) {
      upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
    }
  }

  return ptr;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeFieldTag(upb_Decoder* d, const char* ptr,
                                        uint32_t* field_number,
                                        uint32_t* wire_type) {
  uint32_t tag;
  UPB_ASSERT(ptr < d->input.limit_ptr);
  ptr = upb_WireReader_ReadTag(ptr, &tag, EPS(d));
  *field_number = tag >> 3;
  *wire_type = tag & 7;
  return ptr;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeFieldData(
    upb_Decoder* d, const char* ptr, upb_Message* msg, const upb_MiniTable* mt,
    uint32_t field_number, uint32_t wire_type, const char* start) {
  int op;
  wireval val;

  const upb_MiniTableField* field =
      _upb_Decoder_FindField(d, mt, field_number, wire_type);
  ptr = _upb_Decoder_DecodeWireValue(d, ptr, mt, field, wire_type, &val, &op);

  if (op >= 0) {
    return _upb_Decoder_DecodeKnownField(d, ptr, msg, field, op, &val);
  } else {
    switch (op) {
      case kUpb_DecodeOp_UnknownField:
        return _upb_Decoder_DecodeUnknowns(d, ptr, msg, mt, field_number,
                                           wire_type, val, start);
      case kUpb_DecodeOp_MessageSetItem:
        return upb_Decoder_DecodeMessageSetItem(d, ptr, msg, mt);
      default:
        UPB_UNREACHABLE();
    }
  }
}

static const char* _upb_Decoder_EndMessage(upb_Decoder* d, const char* ptr) {
  d->message_is_done = true;
  return ptr;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeFieldNoFast(upb_Decoder* d, const char* ptr,
                                           upb_Message* msg,
                                           const upb_MiniTable* mt) {
  uint32_t field_number;
  uint32_t wire_type;

  const char* start = ptr;
  ptr = _upb_Decoder_DecodeFieldTag(d, ptr, &field_number, &wire_type);

  if (wire_type == kUpb_WireType_EndGroup) {
    d->end_group = field_number;
    return _upb_Decoder_EndMessage(d, ptr);
  }

  ptr = _upb_Decoder_DecodeFieldData(d, ptr, msg, mt, field_number, wire_type,
                                     start);
  _upb_Decoder_Trace(d, 'M');
  return ptr;
}

UPB_FORCEINLINE
bool _upb_Decoder_TryDecodeMessageFast(upb_Decoder* d, const char** ptr,
                                       upb_Message* msg,
                                       const upb_MiniTable* mt,
                                       uint64_t last_field_index,
                                       uint64_t data) {
#if UPB_FASTTABLE
  if (mt->UPB_PRIVATE(table_mask) == (unsigned char)-1 ||
      (d->options & kUpb_DecodeOption_DisableFastTable)) {
    // Fast table is unavailable or disabled.
    return false;
  }

  uint64_t data2 = upb_DecodeFastData2_PackMask(mt->UPB_PRIVATE(table_mask));
  const char* start =
      UPB_PRIVATE(upb_EpsCopyInputStream_GetInputPtr)(&d->input, *ptr);
  char* trace_next = _upb_Decoder_TraceNext(d);

  upb_FastDecoder_Return ret =
      upb_DecodeFast_Dispatch(d, *ptr, msg, mt, 0, 0, data2);
  *ptr = ret.ptr;

  if (d->message_is_done) {
    // The entire message was successfully parsed fast.
    return true;
  }

  // *ptr now points to the beginning of a field that could not be parsed fast.
  // It's possible that some fields were parsed fast, in which case *ptr will
  // have been updated. However, it's also possible that the very first field
  // encountered could not be parsed fast, in which case *ptr will be unchanged.
  //
  // If the fast decoder consumed any data, it must have emitted at least
  // one 'F' event into the trace buffer (in addition to the 'D' event
  // that is always emitted).
  const char* end =
      UPB_PRIVATE(upb_EpsCopyInputStream_GetInputPtr)(&d->input, *ptr);
  char* trace_end = _upb_Decoder_TracePtr(d);
  UPB_ASSERT(trace_end == NULL || trace_end != trace_next || end == start);
  _upb_Decoder_Trace(d, '<');
#endif
  return false;
}

UPB_FORCEINLINE
const char* _upb_Decoder_DecodeField(upb_Decoder* d, const char* ptr,
                                     upb_Message* msg, const upb_MiniTable* mt,
                                     uint64_t last_field_index, uint64_t data) {
  if (_upb_Decoder_TryDecodeMessageFast(d, &ptr, msg, mt, last_field_index,
                                        data)) {
    return ptr;
  } else if (upb_EpsCopyInputStream_IsDone(EPS(d), &ptr)) {
    return _upb_Decoder_EndMessage(d, ptr);
  }

  return _upb_Decoder_DecodeFieldNoFast(d, ptr, msg, mt);
}

UPB_NOINLINE
static const char* _upb_Decoder_DecodeEmptyMessage(upb_Decoder* d,
                                                   const char* ptr,
                                                   upb_Message* msg) {
  if (upb_EpsCopyInputStream_IsDone(EPS(d), &ptr)) {
    return ptr;
  }

  const char* start = ptr;
  const char* capture_end = ptr;
  upb_EpsCopyCapture capture;
  upb_EpsCopyCapture_Start(&capture, &d->input, start);
  while (!upb_EpsCopyInputStream_IsDone(EPS(d), &ptr)) {
    uint32_t tag;
    capture_end = ptr;
    ptr = upb_WireReader_ReadTag(ptr, &tag, EPS(d));
    if ((tag & 7) == kUpb_WireType_EndGroup) {
      d->end_group = tag >> 3;
      break;
    }
    ptr = _upb_WireReader_SkipValueForceInline(ptr, tag, d->depth, EPS(d));
    capture_end = ptr;
  }
  upb_StringView sv;
  upb_EpsCopyCapture_End(&capture, EPS(d), capture_end, &sv);

  if (sv.size > 0) {
    if (!UPB_PRIVATE(_upb_Message_AddUnknown)(
            msg, sv.data, sv.size, &d->arena,
            _upb_Decoder_GetAddUnknownMode(d, sv.data))) {
      upb_ErrorHandler_ThrowError(d->err, kUpb_DecodeStatus_OutOfMemory);
    }
  }
  return ptr;
}

// When fasttable is enabled, _upb_Decoder_DecodeMessage possibly calls a
// preserve_none function, which forces a spill of all callee-save registers
// registers to the stack in its prologue and restoration in its epilogue, due
// to mismatched calling conventions - the fast decoder (preserve_none) calls
// _upb_Decoder_DecodeMessage (normal) which calls the fast decoder
// (preserve_none). Arm has a lot of callee-save registers in its normal calling
// convention, including a bunch of simd&fp registers that our preserve_none
// caller is probably not actually using. To avoid this cost, all functions in
// the call stack (excluding force-inlined) between the fast decoder's decode
// message function and a recursive call to the fast decoder should use the fast
// decoder's calling convention.
#if UPB_FASTTABLE
UPB_PRESERVE_NONE
#endif
UPB_NOINLINE
const char* _upb_Decoder_DecodeMessage(upb_Decoder* d, const char* ptr,
                                       upb_Message* msg,
                                       const upb_MiniTable* mt) {
  UPB_ASSERT(mt);
  UPB_ASSERT(d->message_is_done == false);

  if (UPB_UNLIKELY(upb_MiniTable_FieldCount(mt) == 0 &&
                   UPB_PRIVATE(_upb_MiniTable_ExtModeBase)(mt) ==
                       kUpb_ExtMode_NonExtendable)) {
    return _upb_Decoder_DecodeEmptyMessage(d, ptr, msg);
  }

  do {
    ptr = _upb_Decoder_DecodeField(d, ptr, msg, mt, 0, 0);
  } while (!d->message_is_done);
  d->message_is_done = false;

  return UPB_UNLIKELY(mt && mt->UPB_PRIVATE(required_count))
             ? _upb_Decoder_CheckRequired(d, ptr, msg, mt)
             : ptr;
}

static upb_DecodeStatus _upb_Decoder_DecodeTop(struct upb_Decoder* d,
                                               const char* buf,
                                               upb_Message* msg,
                                               const upb_MiniTable* m) {
  _upb_Decoder_DecodeMessage(d, buf, msg, m);
  if (d->end_group != DECODE_NOGROUP) return kUpb_DecodeStatus_Malformed;
  if (d->missing_required) return kUpb_DecodeStatus_MissingRequired;
  return kUpb_DecodeStatus_Ok;
}

static upb_DecodeStatus upb_Decoder_Decode(upb_Decoder* const decoder,
                                           const char* const buf,
                                           upb_Message* const msg,
                                           const upb_MiniTable* const m,
                                           upb_Arena* const arena) {
  if (UPB_SETJMP(decoder->err->buf) == 0) {
    decoder->err->code = _upb_Decoder_DecodeTop(decoder, buf, msg, m);
  } else {
    UPB_ASSERT(decoder->err->code != kUpb_DecodeStatus_Ok);
  }

  return upb_Decoder_Destroy(decoder, arena);
}

static uint16_t upb_DecodeOptions_GetMaxDepth(uint32_t options) {
  return options >> 16;
}

uint16_t upb_DecodeOptions_GetEffectiveMaxDepth(uint32_t options) {
  uint16_t max_depth = upb_DecodeOptions_GetMaxDepth(options);
  return max_depth ? max_depth : kUpb_WireFormat_DefaultDepthLimit;
}

upb_DecodeStatus upb_Decode(const char* buf, size_t size, upb_Message* msg,
                            const upb_MiniTable* mt,
                            const upb_ExtensionRegistry* extreg, int options,
                            upb_Arena* arena) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  upb_Decoder decoder;
  upb_ErrorHandler err;
  upb_ErrorHandler_Init(&err);
  buf = upb_Decoder_Init(&decoder, buf, size, extreg, options, arena, &err,
                         NULL, 0);

  return upb_Decoder_Decode(&decoder, buf, msg, mt, arena);
}

upb_DecodeStatus upb_DecodeWithTrace(const char* buf, size_t size,
                                     upb_Message* msg, const upb_MiniTable* mt,
                                     const upb_ExtensionRegistry* extreg,
                                     int options, upb_Arena* arena,
                                     char* trace_buf, size_t trace_size) {
  UPB_ASSERT(!upb_Message_IsFrozen(msg));
  upb_Decoder decoder;
  upb_ErrorHandler err;
  upb_ErrorHandler_Init(&err);
  buf = upb_Decoder_Init(&decoder, buf, size, extreg, options, arena, &err,
                         trace_buf, trace_size);

  return upb_Decoder_Decode(&decoder, buf, msg, mt, arena);
}

upb_DecodeStatus upb_DecodeLengthPrefixed(const char* buf, size_t size,
                                          upb_Message* msg,
                                          size_t* num_bytes_read,
                                          const upb_MiniTable* mt,
                                          const upb_ExtensionRegistry* extreg,
                                          int options, upb_Arena* arena) {
  // To avoid needing to make a Decoder just to decode the initial length,
  // hand-decode the leading varint for the message length here.
  uint64_t msg_len = 0;
  for (size_t i = 0;; ++i) {
    if (i >= size || i > 9) {
      return kUpb_DecodeStatus_Malformed;
    }
    uint64_t b = *buf;
    buf++;
    msg_len += (b & 0x7f) << (i * 7);
    if ((b & 0x80) == 0) {
      *num_bytes_read = i + 1 + msg_len;
      break;
    }
  }

  // If the total number of bytes we would read (= the bytes from the varint
  // plus however many bytes that varint says we should read) is larger then the
  // input buffer then error as malformed.
  if (*num_bytes_read > size) {
    return kUpb_DecodeStatus_Malformed;
  }
  if (msg_len > INT32_MAX) {
    return kUpb_DecodeStatus_Malformed;
  }

  return upb_Decode(buf, msg_len, msg, mt, extreg, options, arena);
}

const char* upb_DecodeStatus_String(upb_DecodeStatus status) {
  switch (status) {
    case kUpb_DecodeStatus_Ok:
      return "Ok";
    case kUpb_DecodeStatus_Malformed:
      return "Wire format was corrupt";
    case kUpb_DecodeStatus_OutOfMemory:
      return "Arena alloc failed";
    case kUpb_DecodeStatus_BadUtf8:
      return "String field had bad UTF-8";
    case kUpb_DecodeStatus_MaxDepthExceeded:
      return "Exceeded upb_DecodeOptions_MaxDepth";
    case kUpb_DecodeStatus_MissingRequired:
      return "Missing required field";
    default:
      return "Unknown decode status";
  }
}

#undef OP_FIXPCK_LG2
#undef OP_VARPCK_LG2

// We encode backwards, to avoid pre-computing lengths (one-pass encode).


#include <setjmp.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Must be last.

UPB_NOINLINE bool _upb_Encoder_AddMapEntryUnknown(
    upb_Message* msg, const upb_MiniTableField* field, upb_Message* ent_msg,
    const upb_MiniTable* entry, upb_Arena* arena) {
  char* buf;
  size_t size;
  upb_EncodeStatus status =
      _upb_Encode(ent_msg, entry, 0, arena, &buf, &size, false);
  if (status != kUpb_EncodeStatus_Ok) {
    return false;
  }
  char delim_buf[2 * kUpb_Encoder_EncodeVarint32MaxSize];
  char* delim_end = delim_buf;
  uint32_t tag =
      ((uint32_t)field->UPB_PRIVATE(number) << 3) | kUpb_WireType_Delimited;
  delim_end = upb_Encoder_EncodeVarint32(tag, delim_end);
  delim_end = upb_Encoder_EncodeVarint32(size, delim_end);
  upb_StringView unknown[] = {
      {delim_buf, delim_end - delim_buf},
      {buf, size},
  };

  return UPB_PRIVATE(_upb_Message_AddUnknownV)(msg, arena, unknown, 2);
}

static uint32_t encode_zz32(int32_t n) {
  return ((uint32_t)n << 1) ^ (n >> 31);
}
static uint64_t encode_zz64(int64_t n) {
  return ((uint64_t)n << 1) ^ (n >> 63);
}

UPB_NORETURN static void encode_err(upb_encstate* e, upb_EncodeStatus s) {
  UPB_ASSERT(s != kUpb_EncodeStatus_Ok);
  e->status = s;
  UPB_LONGJMP(*e->err, 1);
}

UPB_NOINLINE static char* encode_growbuffer(char* ptr, upb_encstate* e,
                                            size_t bytes) {
  ptr = upb_BackAlloc_Grow(&e->alloc, ptr, bytes);
  if (ptr == NULL) {
    encode_err(e, kUpb_EncodeStatus_OutOfMemory);
  }
  return ptr;
}

// Call to ensure that at least `bytes` bytes are available for writing at
// ptr.
UPB_FORCEINLINE
char* encode_reserve(char* ptr, upb_encstate* e, size_t bytes) {
  if (upb_BackAlloc_HasBytes(&e->alloc, ptr, bytes)) {
    return ptr - bytes;
  }
  return encode_growbuffer(ptr, e, bytes);
}

static char* encode_bytes_unchecked(char* ptr, upb_encstate* e,
                                    const void* data, size_t len) {
  if (len == 0) return ptr; /* memcpy() with zero size is UB */
  ptr -= len;
  memcpy(ptr, data, len);
  return ptr;
}

/* Writes the given bytes to the buffer, handling reserve/advance. */
static char* encode_bytes(char* ptr, upb_encstate* e, const void* data,
                          size_t len) {
  if (len == 0) return ptr; /* memcpy() with zero size is UB */
  ptr = encode_reserve(ptr, e, len);
  memcpy(ptr, data, len);
  return ptr;
}

static char* encode_fixed64_unchecked(char* ptr, upb_encstate* e,
                                      uint64_t val) {
  val = upb_BigEndian64(val);
  return encode_bytes_unchecked(ptr, e, &val, sizeof(uint64_t));
}

static char* encode_fixed32_unchecked(char* ptr, upb_encstate* e,
                                      uint32_t val) {
  val = upb_BigEndian32(val);
  return encode_bytes_unchecked(ptr, e, &val, sizeof(uint32_t));
}

#define UPB_PB_VARINT_MAX_LEN 10
#define UPB_PB_VARINT32_MAX_LEN 5

#if UPB_ARM64_ASM
// Each arm64 instruction encodes to 4 bytes, and it takes two intructions
// to process each byte of output, so we branch ahead by (4 + 4) * skip to
// avoid the remaining bytes. When BTI is on, we need to use specific
// "landing pad" instructions, so we pad those with nop to make it a power
// of 2, skipping 16 bytes at each stage instead of 8. This carries some
// overhead especially on in-order cores so they're not included unless
// building with branch protection.
#if UPB_ARM64_BTI_DEFAULT
// BTI is used with jc targets here because we don't control which register will
// be used for addr; if it's x16 or x17 a `br` is treated like a call.
#define UPB_BTI_JC "bti jc\n"
#define UPB_BTI_NOP "nop\n"
#define UPB_BTI_SHIFT_IMM "4\n"
#else
#define UPB_BTI_JC
#define UPB_BTI_NOP
#define UPB_BTI_SHIFT_IMM "3\n"
#endif
UPB_NOINLINE static char* encode_longvarint(char* ptr, upb_encstate* e,
                                            uint64_t val) {
  ptr = encode_reserve(ptr, e, UPB_PB_VARINT_MAX_LEN);
  uint64_t clz;
  __asm__("clz %[cnt], %[val]\n" : [cnt] "=r"(clz) : [val] "r"(val));

  uint32_t skip =
      UPB_PRIVATE(upb_WireWriter_VarintUnusedSizeFromLeadingZeros64)(clz);

  ptr += skip;
  uint64_t addr, mask;
  __asm__ volatile(
      // Formatter keeps merging short lines
      // clang-format off
      "adr %[addr], 0f\n"
      "add %[addr], %[addr], %[cnt], lsl #" UPB_BTI_SHIFT_IMM
      "mov %w[mask], #0x80\n"
      "br %[addr]\n"
      // Work around llvm/llvm-project#47432, where alignment directives inside
      // functions crash the compiler when generating SEH info on windows.
      #ifndef __SEH__
      ".p2align " UPB_BTI_SHIFT_IMM
      #endif
      "0:\n"
      // We don't need addr any more, but we've got the register for our whole
      // assembly block so we'll use it as scratch to store the shift+masked
      // values before storing them.
      // The following stores are unsigned offset stores:
      // strb Wt, [Xn, #imm]
      UPB_BTI_JC
      "orr %[addr], %[mask], %[val], lsr #56\n"
      "strb %w[addr], [%[ptr], #8]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      "orr %[addr], %[mask], %[val], lsr #49\n"
      "strb %w[addr], [%[ptr], #7]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      "orr %[addr], %[mask], %[val], lsr #42\n"
      "strb %w[addr], [%[ptr], #6]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      "orr %[addr], %[mask], %[val], lsr #35\n"
      "strb %w[addr], [%[ptr], #5]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      "orr %[addr], %[mask], %[val], lsr #28\n"
      "strb %w[addr], [%[ptr], #4]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      "orr %w[addr], %w[mask], %w[val], lsr #21\n"
      "strb %w[addr], [%[ptr], #3]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      "orr %w[addr], %w[mask], %w[val], lsr #14\n"
      "strb %w[addr], [%[ptr], #2]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      "orr %w[addr], %w[mask], %w[val], lsr #7\n"
      "strb %w[addr], [%[ptr], #1]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      "orr %w[addr], %w[val], #0x80\n"
      "strb %w[addr], [%[ptr]]\n"
      UPB_BTI_NOP

      UPB_BTI_JC
      // clang-format on
      : [addr] "=&r"(addr), [mask] "=&r"(mask)
      : [val] "r"(val), [ptr] "r"(ptr), [cnt] "r"((uint64_t)skip)
      : "memory");
  uint32_t continuations = UPB_PB_VARINT_MAX_LEN - 1 - skip;
  // msan can't instrument stores in inline assembly
  UPB_PRIVATE(upb_Xsan_MarkInitialized)(ptr, continuations);
  // Encode the final byte after the continuation bytes.
  ptr[continuations] = val >> (7 * continuations);
  return ptr;
}
#undef UPB_BTI_JC
#undef UPB_BTI_NOP
#undef UPB_BTI_SHIFT_IMM
#else
UPB_NOINLINE
static char* encode_longvarint(char* ptr, upb_encstate* e, uint64_t val) {
  ptr = encode_reserve(ptr, e, UPB_PB_VARINT_MAX_LEN);
  size_t len = 0;
  do {
    uint8_t byte = val & 0x7fU;
    val >>= 7;
    if (val) byte |= 0x80U;
    ptr[len++] = byte;
  } while (val);
  char* start = ptr + UPB_PB_VARINT_MAX_LEN - len;
  memmove(start, ptr, len);
  return start;
}
#endif

UPB_FORCEINLINE
char* encode_varint_unchecked(char* ptr, upb_encstate* e, uint64_t val) {
  if (val < 128) {
    --ptr;
    *ptr = val;
    return ptr;
  } else {
    return encode_longvarint(ptr, e, val);
  }
}
UPB_FORCEINLINE
char* encode_varint(char* ptr, upb_encstate* e, uint64_t val) {
  if (val < 128 && upb_BackAlloc_HasBytes(&e->alloc, ptr, 1)) {
    --ptr;
    *ptr = val;
    return ptr;
  } else {
    return encode_longvarint(ptr, e, val);
  }
}

UPB_NOINLINE
char* encode_longlength(char* ptr, upb_encstate* e, uint64_t val) {
  if (val > INT32_MAX) {
    encode_err(e, kUpb_EncodeStatus_MaxSizeExceeded);
  }
  return encode_longvarint(ptr, e, val);
}

UPB_FORCEINLINE
char* encode_length(char* ptr, upb_encstate* e, uint64_t val) {
  if (val < 128 && upb_BackAlloc_HasBytes(&e->alloc, ptr, 1)) {
    --ptr;
    *ptr = val;
    return ptr;
  } else {
    return encode_longlength(ptr, e, val);
  }
}

UPB_FORCEINLINE
char* encode_length_unchecked(char* ptr, upb_encstate* e, uint64_t val) {
  if (val < 128) {
    --ptr;
    *ptr = val;
    return ptr;
  } else {
    return encode_longlength(ptr, e, val);
  }
}

static char* encode_double_unchecked(char* ptr, upb_encstate* e, double d) {
  uint64_t u64;
  UPB_STATIC_ASSERT(sizeof(double) == sizeof(uint64_t), "bad double size");
  memcpy(&u64, &d, sizeof(uint64_t));
  return encode_fixed64_unchecked(ptr, e, u64);
}

static char* encode_float_unchecked(char* ptr, upb_encstate* e, float d) {
  uint32_t u32;
  UPB_STATIC_ASSERT(sizeof(float) == sizeof(uint32_t), "bad float size");
  memcpy(&u32, &d, sizeof(uint32_t));
  return encode_fixed32_unchecked(ptr, e, u32);
}

static char* encode_tag_unchecked(char* ptr, upb_encstate* e,
                                  uint32_t field_number, uint8_t wire_type) {
  return encode_varint_unchecked(ptr, e, (field_number << 3) | wire_type);
}

static char* encode_tag(char* ptr, upb_encstate* e, uint32_t field_number,
                        uint8_t wire_type) {
  return encode_varint(ptr, e, (field_number << 3) | wire_type);
}

static char* encode_fixedarray(char* ptr, upb_encstate* e, const upb_Array* arr,
                               size_t elem_size, uint32_t tag) {
  size_t bytes = upb_Array_Size(arr) * elem_size;
  const char* data = upb_Array_DataPtr(arr);
  const char* arr_ptr = data + bytes - elem_size;

  if (tag || !upb_IsLittleEndian()) {
    while (true) {
      if (elem_size == 4) {
        uint32_t val;
        memcpy(&val, arr_ptr, sizeof(val));
        val = upb_BigEndian32(val);
        ptr = encode_bytes(ptr, e, &val, elem_size);
      } else {
        UPB_ASSERT(elem_size == 8);
        uint64_t val;
        memcpy(&val, arr_ptr, sizeof(val));
        val = upb_BigEndian64(val);
        ptr = encode_bytes(ptr, e, &val, elem_size);
      }

      if (tag) {
        ptr = encode_varint(ptr, e, tag);
      }
      if (arr_ptr == data) break;
      arr_ptr -= elem_size;
    }
    return ptr;
  } else {
    return encode_bytes(ptr, e, data, bytes);
  }
}

/* libpg_query local patch: force-inline encode_scalar/array/field when
 * optimizing, to reduce encoder stack use per message nesting level. Not done
 * in unoptimized builds, where inlining increases stack use instead. See
 * vendor/upb/README. */
#ifdef __OPTIMIZE__
#define PG_QUERY_ENCODE_INLINE UPB_FORCEINLINE
#else
#define PG_QUERY_ENCODE_INLINE static
#endif

PG_QUERY_ENCODE_INLINE char* encode_scalar(char* ptr, upb_encstate* e, const void* field_mem,
                           const upb_MiniTableField* f) {
  // Max size is tag + 10 bytes for max varint or 8 for largest fixed size
#define CASE(ctype, type, wtype, encodeval)                                   \
  {                                                                           \
    const size_t bytes = UPB_PB_VARINT32_MAX_LEN + UPB_PB_VARINT_MAX_LEN;     \
    ptr = encode_reserve(ptr, e, bytes);                                      \
    ptr += bytes;                                                             \
    const ctype val = *(const ctype*)field_mem;                               \
    ptr = encode_##type##_unchecked(ptr, e, encodeval);                       \
    return encode_tag_unchecked(ptr, e, upb_MiniTableField_Number(f), wtype); \
  }

  switch (f->UPB_PRIVATE(descriptortype)) {
    case kUpb_FieldType_Double:
      CASE(double, double, kUpb_WireType_64Bit, val);
    case kUpb_FieldType_Float:
      CASE(float, float, kUpb_WireType_32Bit, val);
    case kUpb_FieldType_Int64:
    case kUpb_FieldType_UInt64:
      CASE(uint64_t, varint, kUpb_WireType_Varint, val);
    case kUpb_FieldType_UInt32:
      CASE(uint32_t, varint, kUpb_WireType_Varint, val);
    case kUpb_FieldType_Int32:
    case kUpb_FieldType_Enum:
      CASE(int32_t, varint, kUpb_WireType_Varint, (int64_t)val);
    case kUpb_FieldType_SFixed64:
    case kUpb_FieldType_Fixed64:
      CASE(uint64_t, fixed64, kUpb_WireType_64Bit, val);
    case kUpb_FieldType_Fixed32:
    case kUpb_FieldType_SFixed32:
      CASE(uint32_t, fixed32, kUpb_WireType_32Bit, val);
    case kUpb_FieldType_Bool:
      CASE(bool, varint, kUpb_WireType_Varint, val);
    case kUpb_FieldType_SInt32:
      CASE(int32_t, varint, kUpb_WireType_Varint, encode_zz32(val));
    case kUpb_FieldType_SInt64:
      CASE(int64_t, varint, kUpb_WireType_Varint, encode_zz64(val));
    case kUpb_FieldType_String:
    case kUpb_FieldType_Bytes: {
      upb_StringView view = *(upb_StringView*)field_mem;
      const size_t max_size =
          UPB_PB_VARINT32_MAX_LEN + UPB_PB_VARINT32_MAX_LEN + view.size;
      ptr = encode_reserve(ptr, e, max_size);
      ptr += max_size;
      ptr = encode_bytes_unchecked(ptr, e, view.data, view.size);
      ptr = encode_length_unchecked(ptr, e, view.size);
      return encode_tag_unchecked(ptr, e, upb_MiniTableField_Number(f),
                                  kUpb_WireType_Delimited);
    }
    case kUpb_FieldType_Group: {
      size_t size;
      upb_Message* submsg = *(upb_Message**)field_mem;
      const upb_MiniTable* subm = upb_MiniTable_GetSubMessageTable(f);
      if (submsg == 0) {
        return ptr;
      }
      if (--e->depth == 0 || stack_is_too_deep()) encode_err(e, kUpb_EncodeStatus_MaxDepthExceeded);
      ptr = encode_tag(ptr, e, upb_MiniTableField_Number(f),
                       kUpb_WireType_EndGroup);
      ptr = encode_message(ptr, e, submsg, subm, &size);
      e->depth++;
      return encode_tag(ptr, e, upb_MiniTableField_Number(f),
                        kUpb_WireType_StartGroup);
    }
    case kUpb_FieldType_Message: {
      size_t size;
      upb_Message* submsg = *(upb_Message**)field_mem;
      const upb_MiniTable* subm = upb_MiniTable_GetSubMessageTable(f);
      if (submsg == 0) {
        return ptr;
      }
      if (--e->depth == 0 || stack_is_too_deep()) encode_err(e, kUpb_EncodeStatus_MaxDepthExceeded);
      ptr = encode_message(ptr, e, submsg, subm, &size);
      e->depth++;
      size_t max_size = UPB_PB_VARINT32_MAX_LEN + UPB_PB_VARINT32_MAX_LEN;
      ptr = encode_reserve(ptr, e, max_size);
      ptr += max_size;
      ptr = encode_length_unchecked(ptr, e, size);
      return encode_tag_unchecked(ptr, e, upb_MiniTableField_Number(f),
                                  kUpb_WireType_Delimited);
    }
    default:
      UPB_UNREACHABLE();
  }
#undef CASE
}

PG_QUERY_ENCODE_INLINE char* encode_array(char* ptr, upb_encstate* e, const upb_Message* msg,
                          const upb_MiniTableField* f) {
  const upb_Array* arr = *UPB_PTR_AT(msg, f->UPB_PRIVATE(offset), upb_Array*);
  bool packed = upb_MiniTableField_IsPacked(f);
  size_t pre_len = upb_BackAlloc_Size(&e->alloc, ptr);

  if (arr == NULL || upb_Array_Size(arr) == 0) {
    return ptr;
  }

#define VARINT_CASE(ctype, encode)                              \
  {                                                             \
    const ctype* start = upb_Array_DataPtr(arr);                \
    const ctype* arr_ptr = start + upb_Array_Size(arr);         \
    if (packed) {                                               \
      do {                                                      \
        arr_ptr--;                                              \
        ptr = encode_varint(ptr, e, encode);                    \
      } while (arr_ptr != start);                               \
    } else {                                                    \
      uint32_t number = upb_MiniTableField_Number(f);           \
      do {                                                      \
        arr_ptr--;                                              \
        ptr = encode_varint(ptr, e, encode);                    \
        ptr = encode_tag(ptr, e, number, kUpb_WireType_Varint); \
      } while (arr_ptr != start);                               \
    }                                                           \
  }                                                             \
  break;

#define TAG(wire_type) (packed ? 0 : (f->UPB_PRIVATE(number) << 3 | wire_type))

  switch (f->UPB_PRIVATE(descriptortype)) {
    case kUpb_FieldType_Double:
      ptr = encode_fixedarray(ptr, e, arr, sizeof(double),
                              TAG(kUpb_WireType_64Bit));
      break;
    case kUpb_FieldType_Float:
      ptr = encode_fixedarray(ptr, e, arr, sizeof(float),
                              TAG(kUpb_WireType_32Bit));
      break;
    case kUpb_FieldType_SFixed64:
    case kUpb_FieldType_Fixed64:
      ptr = encode_fixedarray(ptr, e, arr, sizeof(uint64_t),
                              TAG(kUpb_WireType_64Bit));
      break;
    case kUpb_FieldType_Fixed32:
    case kUpb_FieldType_SFixed32:
      ptr = encode_fixedarray(ptr, e, arr, sizeof(uint32_t),
                              TAG(kUpb_WireType_32Bit));
      break;
    case kUpb_FieldType_Int64:
    case kUpb_FieldType_UInt64:
      VARINT_CASE(uint64_t, *arr_ptr);
    case kUpb_FieldType_UInt32:
      VARINT_CASE(uint32_t, *arr_ptr);
    case kUpb_FieldType_Int32:
    case kUpb_FieldType_Enum:
      VARINT_CASE(int32_t, (int64_t)*arr_ptr);
    case kUpb_FieldType_Bool:
      VARINT_CASE(bool, *arr_ptr);
    case kUpb_FieldType_SInt32:
      VARINT_CASE(int32_t, encode_zz32(*arr_ptr));
    case kUpb_FieldType_SInt64:
      VARINT_CASE(int64_t, encode_zz64(*arr_ptr));
    case kUpb_FieldType_String:
    case kUpb_FieldType_Bytes: {
      const upb_StringView* start = upb_Array_DataPtr(arr);
      const upb_StringView* str_ptr = start + upb_Array_Size(arr);
      do {
        str_ptr--;
        ptr = encode_bytes(ptr, e, str_ptr->data, str_ptr->size);
        ptr = encode_length(ptr, e, str_ptr->size);
        ptr = encode_tag(ptr, e, upb_MiniTableField_Number(f),
                         kUpb_WireType_Delimited);
      } while (str_ptr != start);
      return ptr;
    }
    case kUpb_FieldType_Group: {
      const upb_Message* const* start = upb_Array_DataPtr(arr);
      const upb_Message* const* arr_ptr = start + upb_Array_Size(arr);
      const upb_MiniTable* subm = upb_MiniTable_GetSubMessageTable(f);
      if (--e->depth == 0 || stack_is_too_deep()) encode_err(e, kUpb_EncodeStatus_MaxDepthExceeded);
      do {
        size_t size;
        arr_ptr--;
        ptr = encode_tag(ptr, e, upb_MiniTableField_Number(f),
                         kUpb_WireType_EndGroup);
        ptr = encode_message(ptr, e, *arr_ptr, subm, &size);
        ptr = encode_tag(ptr, e, upb_MiniTableField_Number(f),
                         kUpb_WireType_StartGroup);
      } while (arr_ptr != start);
      e->depth++;
      return ptr;
    }
    case kUpb_FieldType_Message: {
      const upb_Message* const* start = upb_Array_DataPtr(arr);
      const upb_Message* const* arr_ptr = start + upb_Array_Size(arr);
      const upb_MiniTable* subm = upb_MiniTable_GetSubMessageTable(f);
      if (--e->depth == 0 || stack_is_too_deep()) encode_err(e, kUpb_EncodeStatus_MaxDepthExceeded);
      do {
        size_t size;
        arr_ptr--;
        ptr = encode_message(ptr, e, *arr_ptr, subm, &size);
        ptr = encode_length(ptr, e, size);
        ptr = encode_tag(ptr, e, upb_MiniTableField_Number(f),
                         kUpb_WireType_Delimited);
      } while (arr_ptr != start);
      e->depth++;
      return ptr;
    }
  }
#undef VARINT_CASE

  if (packed) {
    ptr = encode_length(ptr, e, upb_BackAlloc_Size(&e->alloc, ptr) - pre_len);
    ptr = encode_tag(ptr, e, upb_MiniTableField_Number(f),
                     kUpb_WireType_Delimited);
  }
  return ptr;
}

static char* encode_mapentry(char* ptr, upb_encstate* e, uint32_t number,
                             const upb_MiniTable* layout,
                             const upb_MapEntry* ent) {
  const upb_MiniTableField* key_field = upb_MiniTable_MapKey(layout);
  const upb_MiniTableField* val_field = upb_MiniTable_MapValue(layout);
  size_t pre_len = upb_BackAlloc_Size(&e->alloc, ptr);
  size_t size;
  ptr = encode_scalar(ptr, e, &ent->v, val_field);
  ptr = encode_scalar(ptr, e, &ent->k, key_field);
  size = upb_BackAlloc_Size(&e->alloc, ptr) - pre_len;
  ptr = encode_length(ptr, e, size);
  ptr = encode_tag(ptr, e, number, kUpb_WireType_Delimited);
  return ptr;
}

static char* encode_map(char* ptr, upb_encstate* e, const upb_Message* msg,
                        const upb_MiniTableField* f) {
  const upb_Map* map = *UPB_PTR_AT(msg, f->UPB_PRIVATE(offset), const upb_Map*);
  const upb_MiniTable* layout = upb_MiniTable_MapEntrySubMessage(f);
  UPB_ASSERT(upb_MiniTable_FieldCount(layout) == 2);

  if (!map || !upb_Map_Size(map)) return ptr;

  if (e->options & kUpb_EncodeOption_Deterministic) {
    _upb_sortedmap sorted;
    _upb_mapsorter_pushmap(
        &e->sorter, layout->UPB_PRIVATE(fields)[0].UPB_PRIVATE(descriptortype),
        map, &sorted);
    upb_MapEntry ent;
    while (_upb_sortedmap_next(&e->sorter, map, &sorted, &ent)) {
      ptr = encode_mapentry(ptr, e, upb_MiniTableField_Number(f), layout, &ent);
    }
    _upb_mapsorter_popmap(&e->sorter, &sorted);
  } else {
    upb_value val;
    if (map->UPB_PRIVATE(is_strtable)) {
      intptr_t iter = UPB_STRTABLE_BEGIN;
      upb_StringView strkey;
      while (upb_strtable_next2(&map->t.strtable, &strkey, &val, &iter)) {
        upb_MapEntry ent;
        _upb_map_fromkey(strkey, &ent.k, map->key_size);
        _upb_map_fromvalue(val, &ent.v, map->val_size);
        ptr =
            encode_mapentry(ptr, e, upb_MiniTableField_Number(f), layout, &ent);
      }
    } else {
      intptr_t iter = UPB_INTTABLE_BEGIN;
      uintptr_t intkey = 0;
      while (upb_inttable_next(&map->t.inttable, &intkey, &val, &iter)) {
        upb_MapEntry ent;
        memcpy(&ent.k, &intkey, map->key_size);
        _upb_map_fromvalue(val, &ent.v, map->val_size);
        ptr =
            encode_mapentry(ptr, e, upb_MiniTableField_Number(f), layout, &ent);
      }
    }
  }
  return ptr;
}

static bool encode_shouldencode(const upb_Message* msg,
                                const upb_MiniTableField* f) {
  if (f->presence == 0) {
    // Proto3 presence or map/array.
    const void* mem = UPB_PTR_AT(msg, f->UPB_PRIVATE(offset), void);
    switch (UPB_PRIVATE(_upb_MiniTableField_GetRep)(f)) {
      case kUpb_FieldRep_1Byte: {
        char ch;
        memcpy(&ch, mem, 1);
        return ch != 0;
      }
      case kUpb_FieldRep_4Byte: {
        uint32_t u32;
        memcpy(&u32, mem, 4);
        return u32 != 0;
      }
      case kUpb_FieldRep_8Byte: {
        uint64_t u64;
        memcpy(&u64, mem, 8);
        return u64 != 0;
      }
      case kUpb_FieldRep_StringView: {
        const upb_StringView* str = (const upb_StringView*)mem;
        return str->size != 0;
      }
      default:
        UPB_UNREACHABLE();
    }
  } else if (UPB_PRIVATE(_upb_MiniTableField_HasHasbit)(f)) {
    // Proto2 presence: hasbit.
    return UPB_PRIVATE(_upb_Message_GetHasbit)(msg, f);
  } else {
    // Field is in a oneof.
    return UPB_PRIVATE(_upb_Message_GetOneofCase)(msg, f) ==
           upb_MiniTableField_Number(f);
  }
}

PG_QUERY_ENCODE_INLINE char* encode_field(char* ptr, upb_encstate* e, const upb_Message* msg,
                          const upb_MiniTableField* field) {
  switch (UPB_PRIVATE(_upb_MiniTableField_Mode)(field)) {
    case kUpb_FieldMode_Array:
      return encode_array(ptr, e, msg, field);
    case kUpb_FieldMode_Map:
      return encode_map(ptr, e, msg, field);
    case kUpb_FieldMode_Scalar:
      return encode_scalar(
          ptr, e, UPB_PTR_AT(msg, field->UPB_PRIVATE(offset), void), field);
    default:
      UPB_UNREACHABLE();
  }
}

static char* encode_msgset_item(char* ptr, upb_encstate* e,
                                const upb_MiniTableExtension* ext,
                                const upb_MessageValue ext_val) {
  size_t size;
  ptr = encode_tag(ptr, e, kUpb_MsgSet_Item, kUpb_WireType_EndGroup);
  ptr = encode_message(ptr, e, ext_val.msg_val,
                       upb_MiniTableExtension_GetSubMessage(ext), &size);
  ptr = encode_varint(ptr, e, size);
  ptr = encode_tag(ptr, e, kUpb_MsgSet_Message, kUpb_WireType_Delimited);
  ptr = encode_varint(ptr, e, upb_MiniTableExtension_Number(ext));
  ptr = encode_tag(ptr, e, kUpb_MsgSet_TypeId, kUpb_WireType_Varint);
  ptr = encode_tag(ptr, e, kUpb_MsgSet_Item, kUpb_WireType_StartGroup);
  return ptr;
}

static char* encode_ext(char* ptr, upb_encstate* e,
                        const upb_MiniTableExtension* ext,
                        upb_MessageValue ext_val, bool is_message_set) {
  if (UPB_UNLIKELY(is_message_set)) {
    ptr = encode_msgset_item(ptr, e, ext, ext_val);
  } else {
    ptr = encode_field(ptr, e, &ext_val.UPB_PRIVATE(ext_msg_val),
                       &ext->UPB_PRIVATE(field));
  }
  return ptr;
}

static char* encode_exts(char* ptr, upb_encstate* e, const upb_MiniTable* m,
                         const upb_Message* msg) {
  if (UPB_PRIVATE(_upb_MiniTable_ExtModeBase)(m) == kUpb_ExtMode_NonExtendable)
    return ptr;

  upb_Message_Internal* in = UPB_PRIVATE(_upb_Message_GetInternal)(msg);
  if (!in) return ptr;

  /* Encode all extensions together. Unlike C++, we do not attempt to keep
   * these in field number order relative to normal fields or even to each
   * other. */
  uintptr_t iter = kUpb_Message_ExtensionBegin;
  const upb_MiniTableExtension* ext;
  upb_MessageValue ext_val;
  if (!UPB_PRIVATE(_upb_Message_NextExtensionReverse)(msg, &ext, &ext_val,
                                                      &iter)) {
    // Message has no extensions.
    return ptr;
  }

  if (e->options & kUpb_EncodeOption_Deterministic) {
    _upb_sortedmap sorted;
    if (!_upb_mapsorter_pushexts(&e->sorter, in, &sorted)) {
      // TODO: b/378744096 - handle alloc failure
    }
    const upb_Extension* ext;
    while (_upb_sortedmap_nextext(&e->sorter, &sorted, &ext)) {
      ptr = encode_ext(ptr, e, ext->ext, ext->data,
                       UPB_PRIVATE(_upb_MiniTable_ExtModeBase)(m) ==
                           kUpb_ExtMode_IsMessageSet);
    }
    _upb_mapsorter_popmap(&e->sorter, &sorted);
  } else {
    do {
      ptr = encode_ext(ptr, e, ext, ext_val,
                       m->UPB_PRIVATE(ext) == kUpb_ExtMode_IsMessageSet);
    } while (UPB_PRIVATE(_upb_Message_NextExtensionReverse)(msg, &ext, &ext_val,
                                                            &iter));
  }
  return ptr;
}

/* the minitable of pg_query's Node message (added for Doltgres) */
extern const upb_MiniTable pg_0query__Node_msg_init;

char* encode_message(char* ptr, upb_encstate* e, const upb_Message* msg,
                     const upb_MiniTable* m, size_t* size) {
  size_t pre_len = upb_BackAlloc_Size(&e->alloc, ptr);

  if (e->options & kUpb_EncodeOption_CheckRequired) {
    if (m->UPB_PRIVATE(required_count)) {
      if (!UPB_PRIVATE(_upb_Message_IsInitializedShallow)(msg, m)) {
        encode_err(e, kUpb_EncodeStatus_MissingRequired);
      }
    }
  }

  if ((e->options & kUpb_EncodeOption_SkipUnknown) == 0) {
    size_t unknown_size = 0;
    uintptr_t iter = kUpb_Message_UnknownBegin;
    upb_StringView unknown;
    // Need to write in reverse order, but iteration is in-order; scan to
    // reserve capacity up front, then write in-order
    while (upb_Message_NextUnknown(msg, &unknown, &iter)) {
      unknown_size += unknown.size;
    }
    if (unknown_size != 0) {
      ptr = encode_reserve(ptr, e, unknown_size);
      char* tmp_ptr = ptr;
      iter = kUpb_Message_UnknownBegin;
      while (upb_Message_NextUnknown(msg, &unknown, &iter)) {
        memcpy(tmp_ptr, unknown.data, unknown.size);
        tmp_ptr += unknown.size;
      }
    }
  }

  ptr = encode_exts(ptr, e, m, msg);

  if (m == &pg_0query__Node_msg_init) {
    /*
     * pg_query's Node message is one oneof over all of its several hundred
     * fields, so it encodes the one field that its case names, found by
     * number, rather than check every field, which cost most of each parse.
     * (A Doltgres patch to the vendored upb.)
     */
    const upb_MiniTableField* f =
        upb_Message_WhichOneof(msg, m, &m->UPB_PRIVATE(fields)[0]);
    if (f) ptr = encode_field(ptr, e, msg, f);
  } else if (upb_MiniTable_FieldCount(m)) {
    const upb_MiniTableField* f =
        &m->UPB_PRIVATE(fields)[m->UPB_PRIVATE(field_count)];
    const upb_MiniTableField* first = &m->UPB_PRIVATE(fields)[0];
    while (f != first) {
      f--;
      if (encode_shouldencode(msg, f)) {
        ptr = encode_field(ptr, e, msg, f);
      }
    }
  }

  *size = upb_BackAlloc_Size(&e->alloc, ptr) - pre_len;
  return ptr;
}

static upb_EncodeStatus upb_Encoder_Encode(char* ptr,
                                           upb_encstate* const encoder,
                                           const upb_Message* const msg,
                                           const upb_MiniTable* const l,
                                           char** const buf, size_t* const size,
                                           bool prepend_len) {
  // Unfortunately we must continue to perform hackery here because there are
  // code paths which blindly copy the returned pointer without bothering to
  // check for errors until much later (b/235839510). So we still set *buf to
  // NULL on error and we still set it to non-NULL on a successful empty result.
  if (UPB_SETJMP(*encoder->err) == 0) {
    size_t encoded_msg_size;
    ptr = encode_message(ptr, encoder, msg, l, &encoded_msg_size);
    if (prepend_len) {
      ptr = encode_length(ptr, encoder, encoded_msg_size);
    }
    *size = upb_BackAlloc_Finish(&encoder->alloc, ptr);
    *buf = ptr;
  } else {
    UPB_ASSERT(encoder->status != kUpb_EncodeStatus_Ok);
    upb_BackAlloc_Abort(&encoder->alloc);
    *buf = NULL;
    *size = 0;
  }
  UPB_PRIVATE(_upb_encstate_destroy)(encoder);
  return encoder->status;
}

upb_EncodeStatus _upb_Encode(const upb_Message* msg, const upb_MiniTable* l,
                             int options, upb_Arena* arena, char** buf,
                             size_t* size, bool prepend_len) {
  upb_encstate e;
  jmp_buf err;

  e.status = kUpb_EncodeStatus_Ok;
  e.err = &err;
  char* ptr = upb_BackAlloc_Init(&e.alloc, arena);
  e.depth = upb_EncodeOptions_GetEffectiveMaxDepth(options);
  e.options = options;
  _upb_mapsorter_init(&e.sorter);

  return upb_Encoder_Encode(ptr, &e, msg, l, buf, size, prepend_len);
}

upb_EncodeStatus UPB_PRIVATE(_upb_Encode_Field)(upb_encstate* e,
                                                const upb_Message* msg,
                                                const upb_MiniTableField* field,
                                                char** buf, size_t* size,
                                                int options) {
  e->options = options;
  e->depth = upb_EncodeOptions_GetEffectiveMaxDepth(options);
  char* ptr = *buf;
  if (encode_shouldencode(msg, field)) {
    ptr = encode_field(ptr, e, msg, field);
  }
  *size = upb_BackAlloc_Finish(&e->alloc, ptr);
  *buf = ptr;
  return e->status;
}

upb_EncodeStatus UPB_PRIVATE(_upb_Encode_Extension)(
    upb_encstate* e, const upb_MiniTableExtension* ext,
    upb_MessageValue ext_val, bool is_message_set, char** buf, size_t* size,
    int options) {
  e->options = options;
  e->depth = upb_EncodeOptions_GetEffectiveMaxDepth(options);
  char* ptr = *buf;
  ptr = encode_ext(ptr, e, ext, ext_val, is_message_set);
  *size = upb_BackAlloc_Finish(&e->alloc, ptr);
  *buf = ptr;
  return e->status;
}


#include <stddef.h>
#include <stdint.h>
#include <string.h>


// Must be last.

char upb_BackAlloc_sentinel;

static size_t upb_BackAlloc_CalcBlockSize(upb_BackAlloc* a, size_t required,
                                          bool* one_off) {
#if UPB_HWASAN
  required = UPB_ALIGN_UP(required, UPB_MALLOC_ALIGN);
#endif
  size_t organic_block_size =
      UPB_PRIVATE(_upb_Arena_NextBlockSize)(a->arena, required, one_off);

  // We want to offer amortized linear time, which means we must grow block
  // sizes exponentially. However, merely allocating a power of 2 is
  // pathological for allocators that use a header with mmap for large
  // contiguous allocations. Instead, we want to allocate based on a power of
  // 2, but request slightly less to leave room for backing allocator
  // metadata. If we had universal size feedback this would not be necessary.
  size_t scaled_block_size = upb_RoundUpToPowerOfTwo(required);

  // Estimated value such that 128 bytes of possible overhead is not
  // significant. 128 bytes should be enough for whatever metadata is needed.
  if (scaled_block_size >= 4096 * 4) {
    scaled_block_size = upb_RoundUpToPowerOfTwo(required + 128) - 128;
  }

  // Scaled block size calculations could overflow, but that's OK as it's
  // unsigned and won't be used if it's less than the organic block size
  if (scaled_block_size > organic_block_size) {
    return UPB_PRIVATE(_upb_Arena_NextBlockSize)(a->arena, scaled_block_size,
                                                 one_off);
  }

  return organic_block_size;
}

static char* upb_BackAlloc_Realloc(upb_BackAlloc* a, char* ptr, size_t n) {
  size_t copy = a->limit - ptr;
  if (SIZE_MAX - copy < n) {
    return NULL;
  }

  bool one_off = false;
  size_t required_block_size = copy + n;
  size_t size = upb_BackAlloc_CalcBlockSize(a, required_block_size, &one_off);

  char* block = UPB_PRIVATE(_upb_Arena_AllocBlock)(a->arena, &size);

  if (!block) {
    return NULL;
  }

  UPB_PRIVATE(_upb_Arena_UpdateGrowthState)(a->arena, required_block_size, size,
                                            one_off);

  char* dst = block + size - copy;
  memcpy(dst, ptr, copy);

  if (a->limit != a->buf) {
    // Dispose of the old block.
    if (a->standalone) {
      // Note: while it would technically be possible to give this block to the
      // arena to use for allocations, this could lead to a lot of garbage
      // blocks that never get used.
      UPB_PRIVATE(_upb_Arena_FreeBlock)(a->arena, a->buf);
    } else {
      UPB_PRIVATE(_upb_Arena_UseBlock)(a->arena, a->buf, a->limit - a->buf);
    }
  }

  a->buf = block;
  a->limit = block + size;
  a->standalone = true;
  return dst - n;
}

char* upb_BackAlloc_Grow(upb_BackAlloc* a, char* ptr, size_t n) {
  if (a->limit == a->buf) {
    // First allocation: try to steal a block.
    size_t size = n;
    char* block = UPB_PRIVATE(_upb_Arena_Steal)(a->arena, &size);
    if (block) {
      UPB_ASSERT(size >= n);
      UPB_ASSERT(a->standalone == false);
      a->buf = block;
      a->limit = block + size;
      return a->limit - n;
    }
  }

  return upb_BackAlloc_Realloc(a, ptr, n);
}




// Must be last.

UPB_NOINLINE
const char* _upb_Decoder_CheckRequired(upb_Decoder* d, const char* ptr,
                                       const upb_Message* msg,
                                       const upb_MiniTable* m) {
  UPB_ASSERT(m->UPB_PRIVATE(required_count));
  if (UPB_UNLIKELY(d->options & kUpb_DecodeOption_CheckRequired)) {
    d->missing_required =
        d->missing_required ||
        !UPB_PRIVATE(_upb_Message_IsInitializedShallow)(msg, m);
  }
  return ptr;
}

// We encode backwards, to avoid pre-computing lengths (one-pass encode).


#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>


// Must be last.

upb_EncodeStatus upb_Encode(const upb_Message* msg, const upb_MiniTable* l,
                            int options, upb_Arena* arena, char** buf,
                            size_t* size) {
  return _upb_Encode(msg, l, options, arena, buf, size, false);
}

upb_EncodeStatus upb_EncodeLengthPrefixed(const upb_Message* msg,
                                          const upb_MiniTable* l, int options,
                                          upb_Arena* arena, char** buf,
                                          size_t* size) {
  return _upb_Encode(msg, l, options, arena, buf, size, true);
}

const char* upb_EncodeStatus_String(upb_EncodeStatus status) {
  switch (status) {
    case kUpb_EncodeStatus_Ok:
      return "Ok";
    case kUpb_EncodeStatus_MissingRequired:
      return "Missing required field";
    case kUpb_EncodeStatus_MaxDepthExceeded:
      return "Max depth exceeded";
    case kUpb_EncodeStatus_OutOfMemory:
      return "Arena alloc failed";
    default:
      return "Unknown encode status";
  }
}


#include <stddef.h>
#include <stdint.h>


// Must be last.

UPB_NOINLINE UPB_PRIVATE(_upb_WireReader_LongVarint)
    UPB_PRIVATE(_upb_WireReader_ReadLongVarint)(
        const char* ptr, uint64_t val, upb_EpsCopyInputStream* stream) {
  for (int i = 1; i < 10; i++) {
    uint64_t byte = (uint8_t)ptr[i];
    val += (byte - 1) << (i * 7);
    if (!(byte & 0x80)) {
      return (UPB_PRIVATE(_upb_WireReader_LongVarint)){ptr + i + 1, val};
    }
  }
  return (UPB_PRIVATE(_upb_WireReader_LongVarint)){
      UPB_PRIVATE(upb_EpsCopyInputStream_ReturnError)(stream), 0};
}

UPB_NOINLINE UPB_PRIVATE(_upb_WireReader_LongVarint)
    UPB_PRIVATE(_upb_WireReader_ReadLongTag)(const char* ptr, uint64_t val,
                                             upb_EpsCopyInputStream* stream) {
  for (int i = 1; i < 5; i++) {
    uint64_t byte = (uint8_t)ptr[i];
    val += (byte - 1) << (i * 7);
    if (!(byte & 0x80)) {
      if (val > UINT32_MAX) break;
      return (UPB_PRIVATE(_upb_WireReader_LongVarint)){ptr + i + 1, val};
    }
  }
  return (UPB_PRIVATE(_upb_WireReader_LongVarint)){
      UPB_PRIVATE(upb_EpsCopyInputStream_ReturnError)(stream), 0};
}

UPB_NOINLINE UPB_PRIVATE(_upb_WireReader_LongVarint)
    UPB_PRIVATE(_upb_WireReader_ReadLongSize)(const char* ptr, uint64_t val,
                                              upb_EpsCopyInputStream* stream) {
  for (int i = 1; i < 5; i++) {
    uint64_t byte = (uint8_t)ptr[i];
    val += (byte - 1) << (i * 7);
    if (!(byte & 0x80)) {
      if (val > INT32_MAX) break;
      return (UPB_PRIVATE(_upb_WireReader_LongVarint)){ptr + i + 1, val};
    }
  }
  return (UPB_PRIVATE(_upb_WireReader_LongVarint)){
      UPB_PRIVATE(upb_EpsCopyInputStream_ReturnError)(stream), 0};
}

const char* UPB_PRIVATE(_upb_WireReader_SkipGroup)(
    const char* ptr, uint32_t tag, int depth_limit,
    upb_EpsCopyInputStream* stream) {
  if (--depth_limit < 0) {
    return UPB_PRIVATE(upb_EpsCopyInputStream_ReturnError)(stream);
  }
  uint32_t end_group_tag = (tag & ~7ULL) | kUpb_WireType_EndGroup;
  while (!upb_EpsCopyInputStream_IsDone(stream, &ptr)) {
    uint32_t tag;
    ptr = upb_WireReader_ReadTag(ptr, &tag, stream);
    if (!ptr) break;
    if (tag == end_group_tag) return ptr;
    ptr = _upb_WireReader_SkipValue(ptr, tag, depth_limit, stream);
    if (!ptr) break;
  }
  // Encountered limit end before end group tag.
  return UPB_PRIVATE(upb_EpsCopyInputStream_ReturnError)(stream);
}


#include <stdint.h>
#include <string.h>


// Must be last.

UPB_NORETURN UPB_NOINLINE void UPB_PRIVATE(
    upb_EpsCopyInputStream_ThrowMalformed)(upb_EpsCopyInputStream* e) {
  e->error = true;
  upb_ErrorHandler_ThrowError(e->err, kUpb_ErrorCode_Malformed);
}

const char* UPB_PRIVATE(upb_EpsCopyInputStream_IsDoneFallback)(
    struct upb_EpsCopyInputStream* e, const char* ptr, int overrun) {
  if (overrun < e->limit) {
    // Need to copy remaining data into patch buffer.
    UPB_ASSERT(overrun < kUpb_EpsCopyInputStream_SlopBytes);
    const char* old_end = ptr;
    const char* new_start = &e->patch[overrun];
    memset(&e->patch[kUpb_EpsCopyInputStream_SlopBytes], 0,
           kUpb_EpsCopyInputStream_SlopBytes);
    memcpy(e->patch, e->end, kUpb_EpsCopyInputStream_SlopBytes);
    ptr = new_start;
    e->end = &e->patch[kUpb_EpsCopyInputStream_SlopBytes];
    e->limit -= kUpb_EpsCopyInputStream_SlopBytes;
    e->limit_ptr = e->end + e->limit;
    UPB_ASSERT(ptr < e->limit_ptr);
    e->input_delta = (uintptr_t)old_end - (uintptr_t)new_start;
    UPB_PRIVATE(upb_EpsCopyInputStream_BoundsChecked)(e);
    return new_start;
  } else {
    UPB_ASSERT(overrun > e->limit);
    return UPB_PRIVATE(upb_EpsCopyInputStream_ReturnError)(e);
  }
}


#include <stddef.h>


// Must be last.

#ifdef __cplusplus
extern "C" {
#endif

size_t upb_ByteSize(const upb_Message* msg, const upb_MiniTable* mt) {
  upb_Arena* arena = upb_Arena_New();
  char* buf;
  size_t res = 0;

  (void)upb_Encode(msg, mt, 0, arena, &buf, &res);

  upb_Arena_Free(arena);
  return res;
}

#ifdef __cplusplus
}  // extern "C"
#endif
// This should #undef all macros #defined in def.inc

#undef UPB_SIZE
#undef UPB_PTR_AT
#undef UPB_SIZEOF_FLEX
#undef UPB_SIZEOF_FLEX_WOULD_OVERFLOW
#undef UPB_MAPTYPE_STRING
#undef UPB_EXPORT
#undef UPB_INLINE
#undef UPB_API
#undef UPBC_API
#undef UPB_API_INLINE
#undef UPB_API_INLINE_IF_NOT_GCC
#undef UPB_ALIGN_UP
#undef UPB_ALIGN_DOWN
#undef UPB_ALIGN_MALLOC
#undef UPB_ALIGN_OF
#undef UPB_ALIGN_AS
#undef UPB_STATIC_ASSERT
#undef UPB_STATIC_ASSERT_CONCAT
#undef UPB_STATIC_ASSERT_CONCAT_IMPL
#undef UPB_LIKELY
#undef UPB_UNLIKELY
#undef UPB_UNPREDICTABLE
#undef UPB_FORCEINLINE
#undef UPB_NOINLINE
#undef UPB_NORETURN
#undef UPB_PRINTF
#undef UPB_NODEREF
#undef UPB_NODISCARD
#undef UPB_MAX
#undef UPB_MIN
#undef UPB_UNUSED
#undef UPB_ASSUME
#undef UPB_ASSERT
#undef UPB_UNREACHABLE
#undef UPB_UNREACHABLE_FAILURE
#undef UPB_PRETTY_FUNCTION
#undef UPB_DEFAULT_MAX_BLOCK_SIZE
#undef UPB_SETJMP
#undef UPB_LONGJMP
#undef UPB_PTRADD
#undef UPB_MUSTTAIL
#undef UPB_PRESERVE_MOST
#undef UPB_PRESERVE_NONE
#undef UPB_FASTTABLE_SUPPORTED
#undef UPB_FASTTABLE_MASK
#undef UPB_FASTTABLE
#undef UPB_FASTTABLE_INIT
#undef UPB_POISON_MEMORY_REGION
#undef UPB_UNPOISON_MEMORY_REGION
#undef UPB_ASAN
#undef UPB_HWASAN
#undef UPB_HWASAN_POISON_TAG
#undef UPB_MSAN
#undef UPB_MALLOC_ALIGN
#undef UPB_TSAN
#undef UPB_DEPRECATED
#undef UPB_GNUC_MIN
#undef UPB_CLANG_MIN
#undef UPB_DESCRIPTOR_UPB_H_FILENAME
#undef UPB_DESC_MINITABLE
#undef UPB_IS_GOOGLE3
#undef UPB_ATOMIC
#undef UPB_USE_C11_ATOMICS
#undef UPB_USE_MSC_ATOMICS
#undef UPB_PRIVATE
#undef UPB_ONLYBITS
#undef UPB_LINKARR_DECLARE
#undef UPB_LINKARR_APPEND
#undef UPB_LINKARR_START
#undef UPB_LINKARR_STOP
#undef UPB_HAS_ATTRIBUTE
#undef UPB_HAS_CPP_ATTRIBUTE
#undef UPB_HAS_BUILTIN
#undef UPB_HAS_EXTENSION
#undef UPB_HAS_FEATURE
#undef UPB_XSAN_MEMBER
#undef UPB_XSAN
#undef UPB_XSAN_STRUCT_SIZE
#undef UPB_ENABLE_REF_CYCLE_CHECKS
#undef UPB_ARM64_ASM
#undef UPB_ARM64_BTI_DEFAULT
#undef UPB_DEPRECATE_AND_INLINE
#undef UPB_MAYBE_ASSUME
#undef UPB_ATTR_CONST
#undef _UPB_STRINGIFY
#undef _UPB_STRINGIFY2
#undef UPB_CONSTRUCTOR
