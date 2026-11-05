/**
 *
 *
 * @module typescript-go/testdata/tests/cases/compiler/pipeOperatorEmit
 *
 * @file      pipeOperatorEmit.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */

// @strict: true
// @target: es2015, es2022, esnext
// @declaration: true

const twice = (n: number) => n * 2;
const add = (a: number, b: number) => a + b;
const values = [1, 2];
const object = { offset: 10, method(n: number) { return this.offset + n; } };

// Each input is captured before its target is evaluated.
export const chained = 1 |> twice |> twice;
export const explicit = ...[1, 2] |> add;
export const expanded = ...[...values] |> add;
export const member = 1 |> (object.method);
export const computed = 1 |> object["method"];
export const optional = 1 |> object?.method;
export const array = [1, 2] |> (items => items.length);
export function defaulted(value = 1 |> twice) { return value; }
export async function awaited() {
    return (await Promise.resolve(1)) |> (await Promise.resolve(twice));
}
export function* yielded() {
    return (yield 1) |> twice;
}
export class C {
    value = 1 |> twice;
    static value = ...[...values] |> add;
}
