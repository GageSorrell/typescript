// Compile with this fork, then run the emitted JavaScript with Node.js.
function equal(actual: unknown, expected: unknown): void {
    if (!Object.is(actual, expected)) throw new Error(`${actual} != ${expected}`);
}

export const length = "hello" |> (value => value.length) |> (value => value * 2);
equal(length, 10);

function describe(count: number, label: string): string {
    return `${label}: ${count}`;
}
equal(...[2, "items"] |> describe, "items: 2");
const tuple = [3, "items"] as const;
equal(...[...tuple] |> describe, "items: 3");
equal([1, 2] |> (values => values.length), 2);

export function sum(a = 7, b = 8): number { return a + b; }
const extra = [1, 2, 99];
const empty: number[] = [];
equal(...[...extra] |> sum, 3);
equal(...[...empty] |> sum, 15);

const receiver = {
    offset: 10,
    add(this: { offset: number }, value: number) { return this.offset + value; },
};
equal(2 |> receiver.add, 12);

const log: string[] = [];
const input = () => { log.push("input"); return 3; };
const target = () => {
    log.push("target");
    return (value: number) => { log.push("call"); return value * 2; };
};
equal(input() |> target(), 6);
equal(log.join(), "input,target,call");

// A promise remains a promise unless the source explicitly awaits it.
const promise = Promise.resolve(2);
equal(promise |> (value => value), promise);

console.log("pipe examples ok");
export {};
