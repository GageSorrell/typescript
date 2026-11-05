/**
 *
 *
 * @module typescript-go/testdata/tests/cases/compiler/pipeOperatorChecking
 *
 * @file      pipeOperatorChecking.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */

// @strict: true
// @target: es2022
// @noEmit: true

declare function identity<T>(value: T): T;
declare function box<T>(value: T): { value: T };
declare function count(value: string): number;
declare function select(value: number): number;
declare function select(value: string): string;
declare function pair(a: number, b: string): { a: number; b: string };
declare function optional(a: number, b?: string): number;
declare function rest(a: number, ...values: string[]): string[];
declare function useCallback(callback: (value: string) => number): number;
declare function apply(callback: (value: string) => number, value: string): number;
declare function factory(): (value: number) => string;

const chained = "hello" |> count |> identity |> box;
const selectedNumber = 1 |> select;
const selectedString = "x" |> select;
const unpacked = ...[1, "x"] |> pair;
const tuple = [1, "x"] as const;
const unpackedTuple = ...[...tuple] |> pair;
const optionalArgument = ...[1] |> optional;
const restArguments = ...[1, "a", "b"] |> rest;
const callback = (value => value.length) |> useCallback;
const unpackedCallback = ...[value => value.length, "x"] |> apply;
const inline = 1 |> (value => value.toFixed());
const inlineMany = ...[1, "x"] |> ((n, s) => n.toFixed() + s.toUpperCase());
const factoryCall = 1 |> factory();
const instantiated = "x" |> identity<string>;
const object = {
    prefix: "value=",
    method(this: { prefix: string }, value: number) { return this.prefix + value; },
};
const receiver = 1 |> object.method;
const parenthesizedReceiver = 1 |> (object.method);
const values = [];
1 |> values.push;
const evolved = values;
