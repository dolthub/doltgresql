// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aggregate

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initVarianceAggs registers the variance and standard deviation functions to the catalog.
func initVarianceAggs() {
	for _, o := range []struct {
		name       string
		sample     bool
		sqrtResult bool
	}{
		{"var_pop", false, false},
		{"var_samp", true, false},
		{"variance", true, false},
		{"stddev_pop", false, true},
		{"stddev_samp", true, true},
		{"stddev", true, true},
	} {
		framework.RegisterAggregateFunction(varianceOverload(o.name, pgtypes.Int16, pgtypes.Numeric, newDecimalVarianceBuffer(int16ToDecimal, o.sample, o.sqrtResult), newDecimalVarianceWindowFunction(int16ToDecimal, o.sample, o.sqrtResult)))
		framework.RegisterAggregateFunction(varianceOverload(o.name, pgtypes.Int32, pgtypes.Numeric, newDecimalVarianceBuffer(int32ToDecimal, o.sample, o.sqrtResult), newDecimalVarianceWindowFunction(int32ToDecimal, o.sample, o.sqrtResult)))
		framework.RegisterAggregateFunction(varianceOverload(o.name, pgtypes.Int64, pgtypes.Numeric, newDecimalVarianceBuffer(int64ToDecimal, o.sample, o.sqrtResult), newDecimalVarianceWindowFunction(int64ToDecimal, o.sample, o.sqrtResult)))
		framework.RegisterAggregateFunction(varianceOverload(o.name, pgtypes.Numeric, pgtypes.Numeric, newDecimalVarianceBuffer(decimalIdentity, o.sample, o.sqrtResult), newDecimalVarianceWindowFunction(decimalIdentity, o.sample, o.sqrtResult)))
		framework.RegisterAggregateFunction(varianceOverload(o.name, pgtypes.Float32, pgtypes.Float64, newFloatVarianceBuffer[float32](o.sample, o.sqrtResult), newFloatVarianceWindowFunction[float32](o.sample, o.sqrtResult)))
		framework.RegisterAggregateFunction(varianceOverload(o.name, pgtypes.Float64, pgtypes.Float64, newFloatVarianceBuffer[float64](o.sample, o.sqrtResult), newFloatVarianceWindowFunction[float64](o.sample, o.sqrtResult)))
	}
}

// int16ToDecimal converts an int16 value to *apd.Decimal, for use as the decimalConvert of a
// decimalVarianceBuffer instantiated over int16 (i.e. var_pop(int2)/var_samp(int2)/etc.).
func int16ToDecimal(v int16) *apd.Decimal { return apd.New(int64(v), 0) }

// int32ToDecimal converts an int32 value to *apd.Decimal, for use as the decimalConvert of a
// decimalVarianceBuffer instantiated over int32 (i.e. var_pop(int4)/var_samp(int4)/etc.).
func int32ToDecimal(v int32) *apd.Decimal { return apd.New(int64(v), 0) }

// varianceOverload builds a single var_pop(...)/var_samp(...)/stddev_pop(...)/etc. overload; see sumOverload,
// which this mirrors.
func varianceOverload(name string, paramType, returnType *pgtypes.DoltgresType, newBuffer framework.NewBufferFn, newWindowFunc framework.NewWindowFunctionFn) framework.Func1Aggregate {
	return framework.Func1Aggregate{
		Function1: framework.Function1{
			Name:   name,
			Return: returnType,
			Parameters: [1]*pgtypes.DoltgresType{
				paramType,
			},
			Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
				return nil, nil
			},
		},
		NewAggBuffer:     newBuffer,
		NewAggWindowFunc: newWindowFunc,
	}
}

// decimalVarianceNumeratorDenominator computes the numerator (n*sumX2 - sumX^2) and denominator (n^2 for
// population variance, n*(n-1) for sample variance) of the sum-of-squares variance formula. Both apd.Decimal
// operations here run under sql.DecimalCtx, whose Precision of 0 disables rounding, so this arithmetic is
// exact (no floating-point-style cancellation risk) regardless of how large sumX/sumX2 get. The caller must
// ensure n is large enough for the requested variant (n>=1 for population, n>=2 for sample) to avoid a
// division by zero.
func decimalVarianceNumeratorDenominator(n int64, sumX, sumX2 *apd.Decimal, sample bool) (numerator, denominator *apd.Decimal, err error) {
	nDec := apd.New(n, 0)
	sumXSquared := new(apd.Decimal)
	if _, err := sql.DecimalCtx.Mul(sumXSquared, sumX, sumX); err != nil {
		return nil, nil, err
	}
	nTimesSumX2 := new(apd.Decimal)
	if _, err := sql.DecimalCtx.Mul(nTimesSumX2, nDec, sumX2); err != nil {
		return nil, nil, err
	}
	numerator = new(apd.Decimal)
	if _, err := sql.DecimalCtx.Sub(numerator, nTimesSumX2, sumXSquared); err != nil {
		return nil, nil, err
	}
	denominator = new(apd.Decimal)
	if sample {
		if _, err := sql.DecimalCtx.Mul(denominator, nDec, apd.New(n-1, 0)); err != nil {
			return nil, nil, err
		}
	} else {
		if _, err := sql.DecimalCtx.Mul(denominator, nDec, nDec); err != nil {
			return nil, nil, err
		}
	}
	return numerator, denominator, nil
}

// numericWeightAndFirstDigit mirrors the weight/firstdigit extraction at the top of Postgres's
// select_div_scale (numeric.c): Postgres numerics internally store their digits in base-10000 "NBASE
// digit" groups aligned to the decimal point, weight is the power-of-10000 position of the most
// significant such group, and firstDigit is that leading group's own value (1-9999).
func numericWeightAndFirstDigit(d *apd.Decimal) (weight, firstDigit int64) {
	if d.IsZero() {
		return 0, 0
	}
	digits := fmt.Sprintf("%d", &d.Coeff)
	numDigits := int64(len(digits))
	e := numDigits - 1 + int64(d.Exponent) // decimal exponent of the leading digit
	w := floorDiv(e, 4)
	k := int64(d.Exponent) - 4*w // digits to pad (k>=0) or trim (k<0) to land on d's own group boundary
	var group string
	if k >= 0 {
		group = digits + strings.Repeat("0", int(k))
	} else if cut := numDigits + k; cut > 0 {
		group = digits[:cut]
	} else {
		group = "0"
	}
	if len(group) > 4 {
		group = group[len(group)-4:]
	}
	fd, _ := strconv.ParseInt(group, 10, 64)
	return w, fd
}

// floorDiv is integer division rounding towards negative infinity (unlike Go's /, which truncates towards
// zero), needed so weights below the decimal point (negative e) group correctly in numericWeightAndFirstDigit.
func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// decimalScale returns the number of digits after the decimal point in d's exact representation, Postgres's
// "dscale" for a NumericVar.
func decimalScale(d *apd.Decimal) int64 {
	if d.Exponent < 0 {
		return int64(-d.Exponent)
	}
	return 0
}

// selectDivScale mirrors Postgres's select_div_scale (numeric.c) exactly: it picks the number of fractional
// digits (rscale) a dividend/divisor division should be rounded to, aiming for at least 16 significant
// digits in the quotient (so numeric division is no less accurate than float8) while never dropping below
// either operand's own number of fractional digits.
func selectDivScale(dividend, divisor *apd.Decimal) int64 {
	w1, fd1 := numericWeightAndFirstDigit(dividend)
	w2, fd2 := numericWeightAndFirstDigit(divisor)
	qweight := w1 - w2
	if fd1 <= fd2 {
		qweight--
	}
	const numericMinSigDigits = 16
	rscale := numericMinSigDigits - qweight*4
	if s := decimalScale(dividend); rscale < s {
		rscale = s
	}
	if s := decimalScale(divisor); rscale < s {
		rscale = s
	}
	if rscale < 0 {
		rscale = 0
	}
	const numericMaxDisplayScale = 1000
	if rscale > numericMaxDisplayScale {
		rscale = numericMaxDisplayScale
	}
	return rscale
}

// quoScale divides dividend/divisor and rounds the result to exactly rscale digits after the decimal point,
// matching Postgres's div_var(..., rscale, round=true) semantics used throughout numeric.c.
func quoScale(dividend, divisor *apd.Decimal, rscale int64) (*apd.Decimal, error) {
	p := dividend.NumDigits() - divisor.NumDigits() + int64(dividend.Exponent) - int64(divisor.Exponent) + 1
	if p < 1 {
		p = 1
	}
	p += rscale + avgGuardDigits
	result := new(apd.Decimal)
	if _, err := sql.DecimalCtx.WithPrecision(uint32(p)).Quo(result, dividend, divisor); err != nil {
		return nil, err
	}
	if _, err := sql.DecimalCtx.WithPrecision(uint32(p)).Quantize(result, result, int32(-rscale)); err != nil {
		return nil, err
	}
	return result, nil
}

// decimalVariance computes the population (sample=false) or sample (sample=true) variance of n values given
// their running sum and sum of squares, using the sum-of-squares formula (n*sumX2 - sumX^2) / divisor, where
// divisor is n^2 for population variance and n*(n-1) for sample variance.
func decimalVariance(n int64, sumX, sumX2 *apd.Decimal, sample bool) (*apd.Decimal, int64, error) {
	numerator, denominator, err := decimalVarianceNumeratorDenominator(n, sumX, sumX2, sample)
	if err != nil {
		return nil, 0, err
	}
	if numerator.Sign() <= 0 {
		return new(apd.Decimal), 0, nil
	}
	rscale := selectDivScale(numerator, denominator)
	variance, err := quoScale(numerator, denominator, rscale)
	if err != nil {
		return nil, 0, err
	}
	return variance, rscale, nil
}

// decimalStdDev takes the square root of a variance already computed by decimalVariance, rounding to the
// same rscale decimalVariance's own division used.
func decimalStdDev(variance *apd.Decimal, rscale int64) (*apd.Decimal, error) {
	if variance.Sign() <= 0 {
		return new(apd.Decimal), nil
	}
	p := variance.NumDigits()
	if variance.Exponent > 0 {
		p += int64(variance.Exponent)
	}
	p = p/2 + 1 + rscale + avgGuardDigits
	res := new(apd.Decimal)
	if _, err := sql.DecimalCtx.WithPrecision(uint32(p)).Sqrt(res, variance); err != nil {
		return nil, err
	}
	if _, err := sql.DecimalCtx.WithPrecision(uint32(p)).Quantize(res, res, int32(-rscale)); err != nil {
		return nil, err
	}
	return res, nil
}

// decimalVarianceBuffer is the GROUP BY buffer for var_pop/var_samp/stddev_pop/stddev_samp/variance/stddev over
// smallint/integer/bigint/numeric input, all of which promote to numeric.
type decimalVarianceBuffer[T int16 | int32 | int64 | *apd.Decimal] struct {
	expr       sql.Expression
	sumX       apd.Decimal
	sumX2      apd.Decimal
	count      int64
	convert    func(T) *apd.Decimal
	sample     bool
	sqrtResult bool
}

var _ sql.AggregationBuffer = (*decimalVarianceBuffer[int64])(nil)

func newDecimalVarianceBuffer[T int16 | int32 | int64 | *apd.Decimal](convert func(T) *apd.Decimal, sample, sqrtResult bool) framework.NewBufferFn {
	return func(exprs []sql.Expression) (sql.AggregationBuffer, error) {
		return &decimalVarianceBuffer[T]{expr: exprs[0], convert: convert, sample: sample, sqrtResult: sqrtResult}, nil
	}
}

func (b *decimalVarianceBuffer[T]) Dispose(ctx *sql.Context) {}

func (b *decimalVarianceBuffer[T]) Eval(ctx *sql.Context) (interface{}, error) {
	minCount := int64(1)
	if b.sample {
		minCount = 2
	}
	if b.count < minCount {
		return nil, nil
	}
	variance, rscale, err := decimalVariance(b.count, &b.sumX, &b.sumX2, b.sample)
	if err != nil {
		return nil, err
	}
	if b.sqrtResult {
		return decimalStdDev(variance, rscale)
	}
	return variance, nil
}

func (b *decimalVarianceBuffer[T]) Update(ctx *sql.Context, row sql.Row) error {
	v, err := b.expr.Eval(ctx, row)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	typedV, ok := v.(T)
	if !ok {
		return errors.Errorf("variance: expected %T, got %T", typedV, v)
	}
	d := b.convert(typedV)
	if _, err = sql.DecimalCtx.Add(&b.sumX, &b.sumX, d); err != nil {
		return err
	}
	dSquared := new(apd.Decimal)
	if _, err = sql.DecimalCtx.Mul(dSquared, d, d); err != nil {
		return err
	}
	if _, err = sql.DecimalCtx.Add(&b.sumX2, &b.sumX2, dSquared); err != nil {
		return err
	}
	b.count++
	return nil
}

// decimalVarianceWindowFunction is the sql.WindowFunction used for var_pop/var_samp/stddev_pop/stddev_samp/
// variance/stddev over smallint/integer/bigint/numeric input within an OVER(...) clause.
type decimalVarianceWindowFunction[T int16 | int32 | int64 | *apd.Decimal] struct {
	framework.WindowFramerState
	expr       sql.Expression
	convert    func(T) *apd.Decimal
	sample     bool
	sqrtResult bool
}

var _ sql.WindowFunction = (*decimalVarianceWindowFunction[int64])(nil)

func newDecimalVarianceWindowFunction[T int16 | int32 | int64 | *apd.Decimal](convert func(T) *apd.Decimal, sample, sqrtResult bool) framework.NewWindowFunctionFn {
	return func(exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
		wf := &decimalVarianceWindowFunction[T]{expr: exprs[0], convert: convert, sample: sample, sqrtResult: sqrtResult}
		if err := wf.BindFramer(window); err != nil {
			return nil, err
		}
		return wf, nil
	}
}

func (w *decimalVarianceWindowFunction[T]) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	b := &decimalVarianceBuffer[T]{expr: w.expr, convert: w.convert, sample: w.sample, sqrtResult: w.sqrtResult}
	for i := interval.Start; i < interval.End; i++ {
		if err := b.Update(ctx, buf[i]); err != nil {
			return nil, err
		}
	}
	return b.Eval(ctx)
}

// floatVariance computes the population (sample=false) or sample (sample=true) variance from a running count n
// and a running sum of squared deviations from the mean (sxx), the latter accumulated incrementally via the
// Youngs-Cramer algorithm in floatVarianceBuffer.Update.
func floatVariance(n, sxx float64, sample bool) float64 {
	divisor := n
	if sample {
		divisor = n - 1
	}
	variance := sxx / divisor
	if variance < 0 {
		return 0
	}
	return variance
}

// floatVarianceBuffer is the GROUP BY buffer for var_pop/var_samp/stddev_pop/stddev_samp/variance/stddev over
// real/double precision input, both of which promote to double precision.
type floatVarianceBuffer[T float32 | float64] struct {
	expr       sql.Expression
	n          float64
	sx         float64
	sxx        float64
	sample     bool
	sqrtResult bool
}

var _ sql.AggregationBuffer = (*floatVarianceBuffer[float64])(nil)

func newFloatVarianceBuffer[T float32 | float64](sample, sqrtResult bool) framework.NewBufferFn {
	return func(exprs []sql.Expression) (sql.AggregationBuffer, error) {
		return &floatVarianceBuffer[T]{expr: exprs[0], sample: sample, sqrtResult: sqrtResult}, nil
	}
}

func (b *floatVarianceBuffer[T]) Dispose(ctx *sql.Context) {}

func (b *floatVarianceBuffer[T]) Eval(ctx *sql.Context) (interface{}, error) {
	minCount := float64(1)
	if b.sample {
		minCount = 2
	}
	if b.n < minCount {
		return nil, nil
	}
	variance := floatVariance(b.n, b.sxx, b.sample)
	if b.sqrtResult {
		return math.Sqrt(variance), nil
	}
	return variance, nil
}

// Update folds the next value into the running count/sum/sum-of-squared-deviations using the Youngs-Cramer
// algorithm.
func (b *floatVarianceBuffer[T]) Update(ctx *sql.Context, row sql.Row) error {
	v, err := b.expr.Eval(ctx, row)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	f, ok := v.(T)
	if !ok {
		return errors.Errorf("variance: expected %T, got %T", f, v)
	}
	fv := float64(f)
	b.n++
	b.sx += fv
	if b.n > 1 {
		tmp := fv*b.n - b.sx
		b.sxx += tmp * tmp / (b.n * (b.n - 1))
	} else if math.IsNaN(fv) || math.IsInf(fv, 0) {
		b.sxx = math.NaN()
	}
	return nil
}

// floatVarianceWindowFunction is the sql.WindowFunction used for var_pop/var_samp/stddev_pop/stddev_samp/
// variance/stddev over real/double precision input within an OVER(...) clause.
type floatVarianceWindowFunction[T float32 | float64] struct {
	framework.WindowFramerState
	expr       sql.Expression
	sample     bool
	sqrtResult bool
}

var _ sql.WindowFunction = (*floatVarianceWindowFunction[float64])(nil)

func newFloatVarianceWindowFunction[T float32 | float64](sample, sqrtResult bool) framework.NewWindowFunctionFn {
	return func(exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
		wf := &floatVarianceWindowFunction[T]{expr: exprs[0], sample: sample, sqrtResult: sqrtResult}
		if err := wf.BindFramer(window); err != nil {
			return nil, err
		}
		return wf, nil
	}
}

func (w *floatVarianceWindowFunction[T]) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	b := &floatVarianceBuffer[T]{expr: w.expr, sample: w.sample, sqrtResult: w.sqrtResult}
	for i := interval.Start; i < interval.End; i++ {
		if err := b.Update(ctx, buf[i]); err != nil {
			return nil, err
		}
	}
	return b.Eval(ctx)
}
