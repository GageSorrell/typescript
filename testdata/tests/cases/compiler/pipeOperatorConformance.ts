// @strict: true
// @target: es2015, es2022, esnext
// @declaration: true

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
