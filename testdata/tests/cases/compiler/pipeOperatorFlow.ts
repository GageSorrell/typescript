/**
 *
 *
 * @module typescript-go/testdata/tests/cases/compiler/pipeOperatorFlow
 *
 * @file      pipeOperatorFlow.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */

// @strict: true
// @target: es2022
// @noEmit: true

declare function isString(value: unknown): value is string;
declare function taggedString(tag: number, value: unknown): value is string;
declare function assertString(value: unknown): asserts value is string;
declare function assert(value: unknown): asserts value;
declare function fail(value: unknown): never;

function predicate(value: string | number) {
    if (value |> isString) {
        const text: string = value;
    } else {
        const count: number = value;
    }
}
function unpackedPredicate(value: unknown) {
    if (...[1, value] |> taggedString) {
        const text: string = value;
    }
}
function assertion(value: unknown) {
    value |> assertString;
    const text: string = value;
}
function truthy(value: string | undefined) {
    value |> assert;
    const text: string = value;
}
function neverReturns(value: unknown): string {
    value |> fail;
}
function immediateEffects() {
    let value: string | number = "x";
    1 |> (_ => { value = 1; });
    const count: number = value;
}
function reassignedTarget() {
    let fn: ((value: number) => number) | undefined;
    return (fn = n => n + 1, 1) |> fn;
}
function stalePredicate(value: string | number) {
    if (value |> (value = 1, isString)) {
        const text: string = value; // Predicate describes the captured value.
    }
}

declare function prefixedGuard(value: unknown, ...rest: unknown[]): value is string;
declare function prefixedAssert(value: unknown, ...rest: unknown[]): asserts value;
declare function shiftedGuard(tag: unknown, value: unknown): value is string;
declare function shiftedAssert(tag: unknown, value: unknown): asserts value;
function stableArrayPrefix(value: string | undefined, rest: unknown[]) {
    if (...[value, ...rest] |> prefixedGuard) {
        const text: string = value;
    }
    ...[value, ...rest] |> prefixedAssert;
    const text: string = value;
}
function shiftedArraySuffix(value: string | undefined, rest: number[]) {
    if (...[...rest, value] |> shiftedGuard) {
        const text: string = value; // A variable-length spread can shift value.
    }
    ...[...rest, value] |> shiftedAssert;
    const text: string = value; // The asserted parameter need not be value.
}
