package checker_test

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestPipeArrayArguments(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"ordinary array":                    `declare function add(a: number, b: number): number; const Foo = [1, 2]; const n: number = ...[...Foo] |> add;`,
		"readonly array":                    `declare function add(a: number, b: number): number; declare const xs: readonly number[]; const n: number = ...[...xs] |> add;`,
		"missing values":                    `declare function add(a: number, b: number): number; const xs: number[] = []; const n: number = ...[...xs] |> add;`,
		"excess values":                     `declare function add(a: number, b: number): number; const xs = [1, 2, 3, 4]; const n: number = ...[...xs] |> add;`,
		"prefix":                            `declare function fn(s: string, a: number, b: number): number; declare const xs: number[]; const n: number = ...["head", ...xs] |> fn;`,
		"suffix":                            `declare function fn(a: number, b: number): number; declare const xs: number[]; const n: number = ...[...xs, 1] |> fn;`,
		"multiple spreads":                  `declare function fn(a: number, b: number): number; declare const a: number[], b: number[]; const n: number = ...[...a, ...b] |> fn;`,
		"optional parameters":               `declare function fn(a?: number, b?: number): number; declare const xs: number[]; const n: number = ...[...xs] |> fn;`,
		"rest parameters":                   `declare function fn(a: number, ...rest: number[]): number; declare const xs: number[]; const n: number = ...[...xs] |> fn;`,
		"overload selection":                `declare function fn(a: number, b: number): number; declare function fn(a: string, b: string): string; declare const xs: string[]; const s: string = ...[...xs] |> fn;`,
		"generic homogeneous":               `declare function fn<T>(a: T, b: T): T; declare const xs: number[]; const n: number = ...[...xs] |> fn;`,
		"generic separate parameters":       `declare function fn<A, B>(a: A, b: B): [A, B]; declare const xs: number[]; const pair: [number, number] = ...[...xs] |> fn;`,
		"generic constraint":                `declare function fn<T extends number>(a: T, b: T): T; declare const xs: number[]; const n: number = ...[...xs] |> fn;`,
		"generic array rest":                `declare function fn<T>(a: T, ...rest: T[]): T; declare const xs: number[]; const n: number = ...[...xs] |> fn;`,
		"generic tuple rest":                `declare function fn<T extends unknown[]>(...args: T): T; declare const xs: number[]; const out: number[] = ...[...xs] |> fn;`,
		"generic rest with prefix":          `declare function fn<T extends unknown[]>(head: string, ...args: T): T; declare const xs: number[]; const out: number[] = ...["head", ...xs] |> fn;`,
		"generic rest after unknown prefix": `declare function fn<T extends unknown[]>(head: number, ...args: T): T; declare const xs: number[]; const out: number[] = ...[...xs] |> fn;`,
		"fixed tuple rest":                  `declare function fn(...args: [number, number]): number; declare const xs: number[]; const n: number = ...[...xs] |> fn;`,
		"variadic tuple rest suffix":        `declare function fn(...args: [...number[], boolean]): string; declare const xs: number[]; const s: string = ...[...xs, true] |> fn;`,
		"generic variadic tuple suffix":     `declare function fn<T>(...args: [...T[], boolean]): T; declare const xs: number[]; const n: number = ...[...xs, true] |> fn;`,
		"optional input tuple":              `declare function fn(a: number | undefined, b: number | undefined): void; declare const tuple: [number?]; declare const xs: number[]; ...[...tuple, ...xs] |> fn;`,
		"fixed tuple and array":             `declare function fn(s: string, a: number, b: number): void; const tuple = ["head"] as const; declare const xs: number[]; ...[...tuple, ...xs] |> fn;`,
		"contextual prefix callback":        `declare function fn(cb: (n: number) => number, ...args: number[]): number; declare const xs: number[]; const n: number = ...[v => v + 1, ...xs] |> fn;`,
		"contextual suffix callback":        `declare function fn(a: (v: number) => number, b: (v: number) => number): void; declare const xs: ((n: number) => number)[]; ...[...xs, v => v + 1] |> fn;`,
		"contextual callback across slots":  `declare function fn(a: (v: string) => number, b: (v: number) => number): void; declare const xs: ((n: string | number) => number)[]; ...[...xs, v => typeof v === "string" ? v.length : v] |> fn;`,
		"ignored unknown excess":            `declare function fn(a: number): number; declare const xs: string[]; const n: number = ...[1, ...xs] |> fn;`,
		"zero receiving positions":          `declare function fn(): number; declare const xs: string[]; const n: number = ...[...xs] |> fn;`,
		"receiver retained":                 `const obj = { value: 1, fn(this: { value: number }, a: number, b: number) { return this.value + a + b; } }; declare const xs: number[]; const n: number = ...[...xs] |> obj.fn;`,
		"generic source array":              `declare function pair<T>(a: T, b: T): [T, T]; function use<T>(xs: T[]) { const out: [T, T] = ...[...xs] |> pair; return out; }`,
		"inline parameters":                 `declare const xs: number[]; const n: number = ...[...xs] |> ((a, b) => a + b);`,
		"inline prefix":                     `declare const xs: number[]; const s: string = ...["head", ...xs] |> ((head, a, b) => head + a.toFixed() + b.toFixed());`,
		"inline rest":                       `declare const xs: number[]; const n: number = ...[...xs] |> ((head, ...tail) => head + tail[0]);`,
		"stable prefix predicate": `declare function guard(value: unknown, ...rest: unknown[]): value is string;
            function use(value: unknown, xs: unknown[]) { if (...[value, ...xs] |> guard) { const s: string = value; } }`,
		"stable prefix assertion": `declare function assert(value: unknown, ...rest: unknown[]): asserts value;
            function use(value: string | undefined, xs: unknown[]) { ...[value, ...xs] |> assert; const s: string = value; }`,
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := newPipeProgram(t, source)
			diags := p.GetSemanticDiagnostics(t.Context(), p.GetSourceFile("/pipe.ts"))
			assert.Equal(t, len(diags), 0, pipeDiagnosticsText(diags))
		})
	}
}

func TestPipeArrayArgumentErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, source string
		code         int32
	}{
		{"incompatible element", `declare function fn(a: number, b: number): void; declare const xs: string[]; ...[...xs] |> fn;`, 2345},
		{"element union must fit every slot", `declare function fn(a: number, b: string): void; declare const xs: (number | string)[]; ...[...xs] |> fn;`, 2345},
		{"spread can reach later slot", `declare function fn(a: number, b: string): void; declare const xs: number[]; ...[...xs, "tail"] |> fn;`, 2345},
		{"suffix can reach first slot", `declare function fn(a: number, b: number | string): void; declare const xs: number[]; ...[...xs, "tail"] |> fn;`, 2345},
		{"multiple spreads overlap", `declare function fn(a: number, b: string): void; declare const a: number[], b: string[]; ...[...a, ...b] |> fn;`, 2345},
		{"unknown element", `declare function fn(a: number): void; declare const xs: unknown[]; ...[...xs] |> fn;`, 2345},
		{"constraint mismatch", `declare function fn<T extends number>(a: T, b: T): T; declare const xs: string[]; ...[...xs] |> fn;`, 2345},
		{"overload mismatch", `declare function fn(a: number, b: number): number; declare function fn(a: string, b: string): string; declare const xs: boolean[]; ...[...xs] |> fn;`, 2769},
		{"tuple suffix checked", `declare function fn(...args: [...number[], boolean]): void; declare const xs: number[]; ...[...xs] |> fn;`, 2345},
		{"rest elements checked", `declare function fn(a: number, ...rest: string[]): void; declare const xs: number[]; ...[...xs] |> fn;`, 2345},
		{"known excess still rejected", `declare function fn(a: number): void; declare const xs: number[]; ...[1, 2, ...xs] |> fn;`, 2554},
		{"fixed tuple arity unchanged", `declare function fn(a: number): void; const tuple = [1, 2] as const; ...[...tuple] |> fn;`, 2554},
		{"plain array stays one argument", `declare function fn(a: number, b?: number): void; const xs = [1, 2]; xs |> fn;`, 2345},
		{"ordinary spread call unchanged", `declare function fn(a: number, b: number): void; const xs = [1, 2]; fn(...xs);`, 2556},
		{"invalid iterable", `declare function fn(a: number): void; ...[...123] |> fn;`, 2488},
		{"ignored expressions checked", `declare function fn(a: number): void; declare const xs: number[]; ...[...xs, missing] |> fn;`, 2304},
		{"shifted predicate argument", `declare function guard(first: unknown, value: unknown): value is string;
            function use(value: unknown, xs: unknown[]) { if (...[...xs, value] |> guard) { const wrong: string = value; } }`, 2322},
		{"shifted assertion argument", `declare function assert(first: unknown, value: unknown): asserts value;
            function use(value: string | undefined, xs: unknown[]) { ...[...xs, value] |> assert; const wrong: string = value; }`, 2322},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			p := newPipeProgram(t, test.source)
			diags := p.GetSemanticDiagnostics(t.Context(), p.GetSourceFile("/pipe.ts"))
			for _, diagnostic := range diags {
				if diagnostic.Code() == test.code {
					return
				}
			}
			t.Fatalf("expected diagnostic %d; got:\n%s", test.code, pipeDiagnosticsText(diags))
		})
	}
}
