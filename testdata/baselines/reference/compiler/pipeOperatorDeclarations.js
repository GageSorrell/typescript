//// [tests/cases/compiler/pipeOperatorDeclarations.ts] ////

//// [pipeOperatorDeclarations.ts]
/**
 *
 *
 * @module typescript-go/testdata/tests/cases/compiler/pipeOperatorDeclarations
 *
 * @file      pipeOperatorDeclarations.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */


declare function count(value: string): number;
declare function identity<T>(value: T): T;
declare function pair(a: number, b: string): { count: number; text: string };

export const size = "abc" |> count;
export const result = ...[1, "x"] |> pair;
export const callbackResult = "abc" |> (value => value.length);
export function generic<T>(value: T) { return value |> identity; }

declare function merge<T>(a: T, b: T): T;
declare function genericPair<T>(a: T, b: T): [T, T];
const values = [1, 2];
export const arrayResult = ...[...values] |> merge;
export function fromArray<T>(values: T[]) { return ...[...values] |> genericPair; }




//// [pipeOperatorDeclarations.d.ts]
/**
 *
 *
 * @module typescript-go/testdata/tests/cases/compiler/pipeOperatorDeclarations
 *
 * @file      pipeOperatorDeclarations.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */
export declare const size: number;
export declare const result: {
    count: number;
    text: string;
};
export declare const callbackResult: number;
export declare function generic<T>(value: T): T;
export declare const arrayResult: number;
export declare function fromArray<T>(values: T[]): [T, T];
