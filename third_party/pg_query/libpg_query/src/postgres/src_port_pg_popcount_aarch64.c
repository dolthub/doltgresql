/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - pg_popcount64
 * - pg_popcount32
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * pg_popcount_aarch64.c
 *	  Holds the AArch64 popcount implementations.
 *
 * Copyright (c) 2025, PostgreSQL Global Development Group
 *
 * IDENTIFICATION
 *	  src/port/pg_popcount_aarch64.c
 *
 *-------------------------------------------------------------------------
 */
#include "c.h"

#include "port/pg_bitutils.h"

#ifdef POPCNT_AARCH64

#include <arm_neon.h>

#ifdef USE_SVE_POPCNT_WITH_RUNTIME_CHECK
#include <arm_sve.h>

#if defined(HAVE_ELF_AUX_INFO) || defined(HAVE_GETAUXVAL)
#include <sys/auxv.h>
/* Ancient glibc releases don't include the HWCAPxxx macros in sys/auxv.h */
#if defined(__linux__) && !defined(HWCAP_SVE)
#include <asm/hwcap.h>
#endif
#endif
#endif

/*
 * The Neon versions are built regardless of whether we are building the SVE
 * versions.
 */
static uint64 pg_popcount_neon(const char *buf, int bytes);
static uint64 pg_popcount_masked_neon(const char *buf, int bytes, bits8 mask);

#ifdef USE_SVE_POPCNT_WITH_RUNTIME_CHECK

/*
 * These are the SVE implementations of the popcount functions.
 */
static uint64 pg_popcount_sve(const char *buf, int bytes);
static uint64 pg_popcount_masked_sve(const char *buf, int bytes, bits8 mask);

/*
 * The function pointers are initially set to "choose" functions.  These
 * functions will first set the pointers to the right implementations (based on
 * what the current CPU supports) and then will call the pointer to fulfill the
 * caller's request.
 */
static uint64 pg_popcount_choose(const char *buf, int bytes);
static uint64 pg_popcount_masked_choose(const char *buf, int bytes, bits8 mask);



#ifdef HAVE_ELF_AUX_INFO
#else
#endif







/*
 * pg_popcount_sve
 *		Returns number of 1 bits in buf
 */


/*
 * pg_popcount_masked_sve
 *		Returns number of 1 bits in buf after applying the mask to each byte
 */


#else							/* USE_SVE_POPCNT_WITH_RUNTIME_CHECK */

/*
 * When the SVE version isn't available, there's no point in using function
 * pointers to vary the implementation.  We instead just make these actual
 * external functions when USE_SVE_POPCNT_WITH_RUNTIME_CHECK is not defined.
 * The compiler should be able to inline the Neon versions here.
 */
uint64
pg_popcount_optimized(const char *buf, int bytes)
{
	return pg_popcount_neon(buf, bytes);
}

uint64
pg_popcount_masked_optimized(const char *buf, int bytes, bits8 mask)
{
	return pg_popcount_masked_neon(buf, bytes, mask);
}

#endif							/* ! USE_SVE_POPCNT_WITH_RUNTIME_CHECK */

/*
 * pg_popcount32
 *		Return number of 1 bits in word
 */
int
pg_popcount32(uint32 word)
{
	return pg_popcount64((uint64) word);
}

/*
 * pg_popcount64
 *		Return number of 1 bits in word
 */
int
pg_popcount64(uint64 word)
{
	/*
	 * For some compilers, __builtin_popcountl() already emits Neon
	 * instructions.  The line below should compile to the same code on those
	 * systems.
	 */
	return vaddv_u8(vcnt_u8(vld1_u8((const uint8 *) &word)));
}

/*
 * pg_popcount_neon
 *		Returns number of 1 bits in buf
 */


/*
 * pg_popcount_masked_neon
 *		Returns number of 1 bits in buf after applying the mask to each byte
 */


#else							/* POPCNT_AARCH64 */

/* prevent linker complaints about empty module */
extern int	pg_popcount_aarch64_dummy_variable;
int			pg_popcount_aarch64_dummy_variable = 0;

#endif							/* ! POPCNT_AARCH64 */
