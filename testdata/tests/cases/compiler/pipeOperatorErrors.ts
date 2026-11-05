/**
 *
 *
 * @module typescript-go/testdata/tests/cases/compiler/pipeOperatorErrors
 *
 * @file      pipeOperatorErrors.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */

// @strict: true
// @target: es2022
// @noEmit: true

declare function stringOnly(value: string): void;
declare function twoNumbers(a: number, b: number): void;
declare function oneNumber(a: number): void;
declare function constrained<T extends string>(value: T): T;
declare function overloaded(value: number): number;
declare function overloaded(value: string): string;
declare function needsThis(this: { value: string }, value: number): void;
declare const maybe: ((value: number) => number) | undefined;

1 |> stringOnly;
1 |> twoNumbers;
...[] |> oneNumber;
...[1, 2] |> oneNumber;
[1, 2] |> oneNumber;
1 |> 2;
1 |> maybe;
1 |> constrained;
true |> overloaded;
1 |> needsThis;

// Unknown-length arrays are accepted by pipe calls.
const values = [1, 2];
...[...values] |> twoNumbers;
