//// [tests/cases/compiler/pipeOperatorEmit.ts] ////

//// [pipeOperatorEmit.ts]
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


//// [pipeOperatorEmit.js]
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
const twice = (n) => n * 2;
const add = (a, b) => a + b;
const values = [1, 2];
const object = { offset: 10, method(n) { return this.offset + n; } };
// Each input is captured before its target is evaluated.
export const chained = (pipeArgs_1 => pipeTarget_1 => (() => void 0).apply.call(pipeTarget_1[1], pipeTarget_1[0], pipeArgs_1))([(pipeArgs_2 => pipeTarget_2 => (() => void 0).apply.call(pipeTarget_2[1], pipeTarget_2[0], pipeArgs_2))([1])([void 0, twice])])([void 0, twice]);
export const explicit = (pipeArgs_3 => pipeTarget_3 => (() => void 0).apply.call(pipeTarget_3[1], pipeTarget_3[0], pipeArgs_3))([1, 2])([void 0, add]);
export const expanded = (pipeArgs_4 => pipeTarget_4 => (() => void 0).apply.call(pipeTarget_4[1], pipeTarget_4[0], pipeArgs_4))([...values])([void 0, add]);
export const member = (pipeArgs_5 => pipeTarget_5 => (() => void 0).apply.call(pipeTarget_5[1], pipeTarget_5[0], pipeArgs_5))([1])((pipeReceiver_1 => [pipeReceiver_1, pipeReceiver_1.method])(object));
export const computed = (pipeArgs_6 => pipeTarget_6 => (() => void 0).apply.call(pipeTarget_6[1], pipeTarget_6[0], pipeArgs_6))([1])(((pipeReceiver_2, pipeKey_1) => [pipeReceiver_2, pipeReceiver_2[pipeKey_1]])(object, "method"));
export const optional = (pipeArgs_7 => pipeTarget_7 => (() => void 0).apply.call(pipeTarget_7[1], pipeTarget_7[0], pipeArgs_7))([1])((pipeReference_1 => pipeReference_1 === void 0 || (pipeReference_1[1] === null || pipeReference_1[1] === void 0) ? void 0 : [pipeReference_1[1], pipeReference_1[1].method])([void 0, object]));
export const array = (pipeArgs_8 => pipeTarget_8 => (() => void 0).apply.call(pipeTarget_8[1], pipeTarget_8[0], pipeArgs_8))([[1, 2]])([void 0, (items => items.length)]);
export function defaulted(value = (pipeArgs_9 => pipeTarget_9 => (() => void 0).apply.call(pipeTarget_9[1], pipeTarget_9[0], pipeArgs_9))([1])([void 0, twice])) { return value; }
export async function awaited() {
    return (pipeArgs_10 => pipeTarget_10 => (() => void 0).apply.call(pipeTarget_10[1], pipeTarget_10[0], pipeArgs_10))([(await Promise.resolve(1))])([void 0, (await Promise.resolve(twice))]);
}
export function* yielded() {
    return (pipeArgs_11 => pipeTarget_11 => (() => void 0).apply.call(pipeTarget_11[1], pipeTarget_11[0], pipeArgs_11))([(yield 1)])([void 0, twice]);
}
export class C {
    value = (pipeArgs_12 => pipeTarget_12 => (() => void 0).apply.call(pipeTarget_12[1], pipeTarget_12[0], pipeArgs_12))([1])([void 0, twice]);
    static value = (pipeArgs_13 => pipeTarget_13 => (() => void 0).apply.call(pipeTarget_13[1], pipeTarget_13[0], pipeArgs_13))([...values])([void 0, add]);
}


//// [pipeOperatorEmit.d.ts]
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
export declare const chained: number;
export declare const explicit: number;
export declare const expanded: number;
export declare const member: number;
export declare const computed: number;
export declare const optional: number;
export declare const array: number;
export declare function defaulted(value?: number): number;
export declare function awaited(): Promise<number>;
export declare function yielded(): Generator<number, number, number>;
export declare class C {
    value: number;
    static value: number;
}
