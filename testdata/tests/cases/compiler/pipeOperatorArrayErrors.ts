/**
 *
 *
 * @module typescript-go/testdata/tests/cases/compiler/pipeOperatorArrayErrors
 *
 * @file      pipeOperatorArrayErrors.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */

// @strict: true
// @target: es2022
// @noEmit: true

declare function numbers(a: number, b: number): void;
declare function one(a: number): void;
declare function heterogeneous(a: number, b: string): void;
declare function tailCanBeEither(a: number, b: number | string): void;
declare function stringRest(a: number, ...rest: string[]): void;
declare function suffix(...args: [...number[], boolean]): void;
declare function select(a: number, b: number): number;
declare function select(a: string, b: string): string;
declare function constrained<T extends number>(a: T, b: T): T;
declare const ns: number[];
declare const ss: string[];
declare const bs: boolean[];
declare const us: unknown[];
declare const mixed: (number | string)[];

...[...ss] |> numbers;
...[...us] |> numbers;
...[...mixed] |> heterogeneous;
...[...ns, "tail"] |> heterogeneous;
...[...ns, "tail"] |> tailCanBeEither;
...[...ns, ...ss] |> heterogeneous;
...[...ns] |> stringRest;
...[...ns] |> suffix;
...[...ss] |> constrained;
...[...bs] |> select;
...[1, 2, ...ns] |> one;
const fixed = [1, 2] as const;
...[...fixed] |> one;
ns |> one;
numbers(...ns); // Ordinary calls retain TypeScript's spread restriction.
