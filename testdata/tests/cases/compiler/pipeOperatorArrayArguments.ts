/**
 *
 *
 * @module typescript-go/testdata/tests/cases/compiler/pipeOperatorArrayArguments
 *
 * @file      pipeOperatorArrayArguments.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */

// @strict: true
// @target: es2022
// @noEmit: true

declare function add(a: number, b: number): number;
declare function prefix(head: string, a: number, b: number): number;
declare function optional(a?: number, b?: number): number;
declare function rest(a: number, ...args: number[]): number;
declare function select(a: number, b: number): number;
declare function select(a: string, b: string): string;
declare function merge<T>(a: T, b: T): T;
declare function pair<A, B>(a: A, b: B): [A, B];
declare function capture<T extends unknown[]>(...args: T): T;
declare function tupleRest(...args: [number, number]): number;
declare function variadic<T>(...args: [...T[], boolean]): T;
declare function callback(cb: (n: number) => number, ...args: number[]): number;

const Foo = [1, 2];
declare const readonlyValues: readonly number[];
declare const strings: string[];
const empty: number[] = [];
const ordinary = ...[...Foo] |> add;
const readonlyInput = ...[...readonlyValues] |> add;
const possiblyMissing = ...[...empty] |> add;
const possiblyExcess = ...[...[1, 2, 3, 4]] |> add;
const mixedPrefix = ...["head", ...Foo] |> prefix;
const mixedSuffix = ...[...Foo, 1] |> add;
const multiple = ...[...Foo, ...readonlyValues] |> add;
const optionalResult = ...[...Foo] |> optional;
const restResult = ...[...Foo] |> rest;
const selected = ...[...strings] |> select;
const generic = ...[...Foo] |> merge;
const separate = ...[...Foo] |> pair;
const captured = ...[...Foo] |> capture;
const fixedRestTuple = ...[...Foo] |> tupleRest;
const variadicSuffix = ...[...Foo, true] |> variadic;
const contextual = ...[n => n.toFixed().length, ...Foo] |> callback;
const inline = ...[...Foo] |> ((a, b) => a.toFixed() + b.toFixed());
const inlineRest = ...[...Foo] |> ((head, ...tail) => head + tail[0]);
declare const optionalTuple: [number?];
declare function acceptsOptional(a: number | undefined, b: number | undefined): void;
...[...optionalTuple, ...Foo] |> acceptsOptional;

function genericSource<T>(values: T[]) {
    return ...[...values] |> pair;
}
