//// [tests/cases/compiler/pipeOperatorConformance.ts] ////

//// [pipeOperatorConformance.ts]
function equal(actual: unknown, expected: unknown): void {
    if (!Object.is(actual, expected)) throw new Error(`${actual} != ${expected}`);
}

function identity<T>(value: T): T { return value; }
function box<T>(value: T): { value: T } { return { value }; }
function select(value: number): number;
function select(value: string): string;
function select(value: number | string) { return value; }
function pair(count: number, text: string) { return { count, text }; }
function sum(a: number = 7, b: number = 8) { return a + b; }
function join(head: string, ...tail: number[]) { return head + tail.join(","); }
function callback(fn: (value: string) => number, value: string) { return fn(value); }
function capture<T extends unknown[]>(...args: T): T { return args; }

export const chain = "hello" |> (value => value.length) |> identity |> box;
export const numberOverload = 1 |> select;
export const stringOverload = "x" |> select;
export const explicit = ...[2, "two"] |> pair;
const tuple = [3, "three"] as const;
export const fixedTuple = ...[...tuple] |> pair;
export const genericTuple = ...[1, "x", true] |> capture;
export const arrayValue = [1, 2] |> (values => values.length);
const values = [1, 2, 99];
const empty: number[] = [];
export const arrayArguments = ...[...values] |> sum;
export const missingArguments = ...[...empty] |> sum;
export const optionalArgument = ...[1] |> sum;
export const restArguments = ...["n=", ...values] |> join;
export const contextualCallback = ...[value => value.length, "hello"] |> callback;
const object = {
    offset: 10,
    method(this: { offset: number }, n: number) { return this.offset + n; },
};
export const receiver = 2 |> object.method;
export const computedReceiver = 3 |> object["method"];

equal(chain.value, box(identity("hello".length)).value);
equal(numberOverload, select(1));
equal(stringOverload, select("x"));
equal(explicit.text, pair(2, "two").text);
equal(fixedTuple.count, pair(...tuple).count);
equal(genericTuple.join(), capture(1, "x", true).join());
equal(arrayValue, [1, 2].length);
equal(arrayArguments, sum(values[0], values[1]));
equal(missingArguments, sum());
equal(optionalArgument, sum(1));
equal(restArguments, join("n=", ...values));
equal(contextualCallback, callback(value => value.length, "hello"));
equal(receiver, object.method(2));
equal(computedReceiver, object["method"](3));

function isText(value: unknown): value is string { return typeof value === "string"; }
function assertText(value: unknown): asserts value is string {
    if (!isText(value)) throw new Error("expected text");
}
export function predicate(value: string | number) {
    if (value |> isText) return value.toUpperCase();
    return value.toFixed();
}
export function assertion(value: unknown) {
    value |> assertText;
    return value.toUpperCase();
}
equal(predicate("hello"), "HELLO");
equal(predicate(2), "2");
equal(assertion("hello"), "HELLO");

// Captures must survive recursive invocations in the target getter.
let entered = false;
let gets = 0;
let inputs = 0;
let calls = 0;
const recursive = {
    get method() {
        gets++;
        if (!entered) {
            entered = true;
            equal(20 |> recursive.method, 30);
        }
        return function(this: typeof recursive, value: number) {
            calls++;
            equal(this, recursive);
            return value + 10;
        };
    },
};
equal((inputs++, 2) |> recursive.method, 12);
equal(inputs, 1);
equal(gets, 2);
equal(calls, 2);

export function defaulted(value = 2 |> identity) { return value; }
export class Fields {
    value = 2 |> (n => n + 1);
    static value = 3 |> (n => n * 2);
}
equal(defaulted(), 2);
equal(new Fields().value, 3);
equal(Fields.value, 6);

export function* yielded(): Generator<string, number, number | ((n: number) => number)> {
    const input = (yield "input") as number;
    return input |> ((yield "target") as (n: number) => number);
}
const generator = yielded();
equal(generator.next().value, "input");
equal(generator.next(3).value, "target");
equal(generator.next(n => n * 2).value, 6);

export async function awaited() {
    const log: string[] = [];
    const input = async () => { log.push("input"); return 3; };
    const target = async () => {
        log.push("target");
        return (n: number) => { log.push("call"); return n * 2; };
    };
    const result = (await input()) |> (await target());
    equal(log.join(), "input,target,call");
    return result;
}
awaited().then(result => {
    equal(result, 6);
    console.log("pipe conformance ok");
});


//// [pipeOperatorConformance.js]
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
function equal(actual, expected) {
    if (!Object.is(actual, expected))
        throw new Error(`${actual} != ${expected}`);
}
function identity(value) { return value; }
function box(value) { return { value }; }
function select(value) { return value; }
function pair(count, text) { return { count, text }; }
function sum(a = 7, b = 8) { return a + b; }
function join(head, ...tail) { return head + tail.join(","); }
function callback(fn, value) { return fn(value); }
function capture(...args) { return args; }
export const chain = (pipeArgs_1 => pipeTarget_1 => (() => void 0).apply.call(pipeTarget_1[1], pipeTarget_1[0], pipeArgs_1))([(pipeArgs_2 => pipeTarget_2 => (() => void 0).apply.call(pipeTarget_2[1], pipeTarget_2[0], pipeArgs_2))([(pipeArgs_3 => pipeTarget_3 => (() => void 0).apply.call(pipeTarget_3[1], pipeTarget_3[0], pipeArgs_3))(["hello"])([void 0, (value => value.length)])])([void 0, identity])])([void 0, box]);
export const numberOverload = (pipeArgs_4 => pipeTarget_4 => (() => void 0).apply.call(pipeTarget_4[1], pipeTarget_4[0], pipeArgs_4))([1])([void 0, select]);
export const stringOverload = (pipeArgs_5 => pipeTarget_5 => (() => void 0).apply.call(pipeTarget_5[1], pipeTarget_5[0], pipeArgs_5))(["x"])([void 0, select]);
export const explicit = (pipeArgs_6 => pipeTarget_6 => (() => void 0).apply.call(pipeTarget_6[1], pipeTarget_6[0], pipeArgs_6))([2, "two"])([void 0, pair]);
const tuple = [3, "three"];
export const fixedTuple = (pipeArgs_7 => pipeTarget_7 => (() => void 0).apply.call(pipeTarget_7[1], pipeTarget_7[0], pipeArgs_7))([...tuple])([void 0, pair]);
export const genericTuple = (pipeArgs_8 => pipeTarget_8 => (() => void 0).apply.call(pipeTarget_8[1], pipeTarget_8[0], pipeArgs_8))([1, "x", true])([void 0, capture]);
export const arrayValue = (pipeArgs_9 => pipeTarget_9 => (() => void 0).apply.call(pipeTarget_9[1], pipeTarget_9[0], pipeArgs_9))([[1, 2]])([void 0, (values => values.length)]);
const values = [1, 2, 99];
const empty = [];
export const arrayArguments = (pipeArgs_10 => pipeTarget_10 => (() => void 0).apply.call(pipeTarget_10[1], pipeTarget_10[0], pipeArgs_10))([...values])([void 0, sum]);
export const missingArguments = (pipeArgs_11 => pipeTarget_11 => (() => void 0).apply.call(pipeTarget_11[1], pipeTarget_11[0], pipeArgs_11))([...empty])([void 0, sum]);
export const optionalArgument = (pipeArgs_12 => pipeTarget_12 => (() => void 0).apply.call(pipeTarget_12[1], pipeTarget_12[0], pipeArgs_12))([1])([void 0, sum]);
export const restArguments = (pipeArgs_13 => pipeTarget_13 => (() => void 0).apply.call(pipeTarget_13[1], pipeTarget_13[0], pipeArgs_13))(["n=", ...values])([void 0, join]);
export const contextualCallback = (pipeArgs_14 => pipeTarget_14 => (() => void 0).apply.call(pipeTarget_14[1], pipeTarget_14[0], pipeArgs_14))([value => value.length, "hello"])([void 0, callback]);
const object = {
    offset: 10,
    method(n) { return this.offset + n; },
};
export const receiver = (pipeArgs_15 => pipeTarget_15 => (() => void 0).apply.call(pipeTarget_15[1], pipeTarget_15[0], pipeArgs_15))([2])((pipeReceiver_1 => [pipeReceiver_1, pipeReceiver_1.method])(object));
export const computedReceiver = (pipeArgs_16 => pipeTarget_16 => (() => void 0).apply.call(pipeTarget_16[1], pipeTarget_16[0], pipeArgs_16))([3])(((pipeReceiver_2, pipeKey_1) => [pipeReceiver_2, pipeReceiver_2[pipeKey_1]])(object, "method"));
equal(chain.value, box(identity("hello".length)).value);
equal(numberOverload, select(1));
equal(stringOverload, select("x"));
equal(explicit.text, pair(2, "two").text);
equal(fixedTuple.count, pair(...tuple).count);
equal(genericTuple.join(), capture(1, "x", true).join());
equal(arrayValue, [1, 2].length);
equal(arrayArguments, sum(values[0], values[1]));
equal(missingArguments, sum());
equal(optionalArgument, sum(1));
equal(restArguments, join("n=", ...values));
equal(contextualCallback, callback(value => value.length, "hello"));
equal(receiver, object.method(2));
equal(computedReceiver, object["method"](3));
function isText(value) { return typeof value === "string"; }
function assertText(value) {
    if (!isText(value))
        throw new Error("expected text");
}
export function predicate(value) {
    if ((pipeArgs_17 => pipeTarget_17 => (() => void 0).apply.call(pipeTarget_17[1], pipeTarget_17[0], pipeArgs_17))([value])([void 0, isText]))
        return value.toUpperCase();
    return value.toFixed();
}
export function assertion(value) {
    (pipeArgs_18 => pipeTarget_18 => (() => void 0).apply.call(pipeTarget_18[1], pipeTarget_18[0], pipeArgs_18))([value])([void 0, assertText]);
    return value.toUpperCase();
}
equal(predicate("hello"), "HELLO");
equal(predicate(2), "2");
equal(assertion("hello"), "HELLO");
// Captures must survive recursive invocations in the target getter.
let entered = false;
let gets = 0;
let inputs = 0;
let calls = 0;
const recursive = {
    get method() {
        gets++;
        if (!entered) {
            entered = true;
            equal((pipeArgs_19 => pipeTarget_19 => (() => void 0).apply.call(pipeTarget_19[1], pipeTarget_19[0], pipeArgs_19))([20])((pipeReceiver_3 => [pipeReceiver_3, pipeReceiver_3.method])(recursive)), 30);
        }
        return function (value) {
            calls++;
            equal(this, recursive);
            return value + 10;
        };
    },
};
equal((pipeArgs_20 => pipeTarget_20 => (() => void 0).apply.call(pipeTarget_20[1], pipeTarget_20[0], pipeArgs_20))([(inputs++, 2)])((pipeReceiver_4 => [pipeReceiver_4, pipeReceiver_4.method])(recursive)), 12);
equal(inputs, 1);
equal(gets, 2);
equal(calls, 2);
export function defaulted(value = (pipeArgs_21 => pipeTarget_21 => (() => void 0).apply.call(pipeTarget_21[1], pipeTarget_21[0], pipeArgs_21))([2])([void 0, identity])) { return value; }
export class Fields {
    constructor() {
        this.value = (pipeArgs_22 => pipeTarget_22 => (() => void 0).apply.call(pipeTarget_22[1], pipeTarget_22[0], pipeArgs_22))([2])([void 0, (n => n + 1)]);
    }
}
Fields.value = (pipeArgs_23 => pipeTarget_23 => (() => void 0).apply.call(pipeTarget_23[1], pipeTarget_23[0], pipeArgs_23))([3])([void 0, (n => n * 2)]);
equal(defaulted(), 2);
equal(new Fields().value, 3);
equal(Fields.value, 6);
export function* yielded() {
    const input = (yield "input");
    return (pipeArgs_24 => pipeTarget_24 => (() => void 0).apply.call(pipeTarget_24[1], pipeTarget_24[0], pipeArgs_24))([input])([void 0, (yield "target")]);
}
const generator = yielded();
equal(generator.next().value, "input");
equal(generator.next(3).value, "target");
equal(generator.next(n => n * 2).value, 6);
export function awaited() {
    return __awaiter(this, void 0, void 0, function* () {
        const log = [];
        const input = () => __awaiter(this, void 0, void 0, function* () { log.push("input"); return 3; });
        const target = () => __awaiter(this, void 0, void 0, function* () {
            log.push("target");
            return (n) => { log.push("call"); return n * 2; };
        });
        const result = (pipeArgs_25 => pipeTarget_25 => (() => void 0).apply.call(pipeTarget_25[1], pipeTarget_25[0], pipeArgs_25))([(yield input())])([void 0, (yield target())]);
        equal(log.join(), "input,target,call");
        return result;
    });
}
awaited().then(result => {
    equal(result, 6);
    console.log("pipe conformance ok");
});


//// [pipeOperatorConformance.d.ts]
export declare const chain: {
    value: number;
};
export declare const numberOverload: number;
export declare const stringOverload: string;
export declare const explicit: {
    count: number;
    text: string;
};
export declare const fixedTuple: {
    count: number;
    text: string;
};
export declare const genericTuple: [number, string, boolean];
export declare const arrayValue: number;
export declare const arrayArguments: number;
export declare const missingArguments: number;
export declare const optionalArgument: number;
export declare const restArguments: string;
export declare const contextualCallback: number;
export declare const receiver: number;
export declare const computedReceiver: number;
export declare function predicate(value: string | number): string;
export declare function assertion(value: unknown): string;
export declare function defaulted(value?: number): number;
export declare class Fields {
    value: number;
    static value: number;
}
export declare function yielded(): Generator<string, number, number | ((n: number) => number)>;
export declare function awaited(): Promise<number>;
