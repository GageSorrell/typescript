/**
 * @module @typescript/native-preview/test/pipe.test
 *
 * @file      pipe.test.ts
 * @author    Gage Sorrell <gage@sorrell.sh>
 * @copyright (c) 2026 Gage Sorrell
 * @license   MIT
 */

import { SyntaxKind } from "@typescript/native-preview/unstable/ast";
import {
    isAssignmentOperatorOrHigher,
    isBinaryOperator,
} from "@typescript/native-preview/unstable/ast/is";
import {
    createScanner,
    stringToToken,
    tokenToString,
} from "@typescript/native-preview/unstable/ast/scanner";
import assert from "node:assert/strict";
import { test } from "node:test";

test("pipe token matches the compiler token and binary operator classification", () => {
    assert.equal(tokenToString(SyntaxKind.BarGreaterThanToken), "|>");
    assert.equal(stringToToken("|>"), SyntaxKind.BarGreaterThanToken);
    assert.ok(isBinaryOperator(SyntaxKind.BarGreaterThanToken));
    assert.ok(isAssignmentOperatorOrHigher(SyntaxKind.BarGreaterThanToken));
});

test("pipe scanning preserves existing bar operators and token ranges", () => {
    const scanner = createScanner(true, undefined, 'value |> fn | > || |= ||= "|>"');
    const kinds = [
        SyntaxKind.Identifier,
        SyntaxKind.BarGreaterThanToken,
        SyntaxKind.Identifier,
        SyntaxKind.BarToken,
        SyntaxKind.GreaterThanToken,
        SyntaxKind.BarBarToken,
        SyntaxKind.BarEqualsToken,
        SyntaxKind.BarBarEqualsToken,
        SyntaxKind.StringLiteral,
    ];
    for (const kind of kinds) {
        assert.equal(scanner.scan(), kind);
        if (kind === SyntaxKind.BarGreaterThanToken) {
            assert.equal(scanner.getTokenText(), "|>");
            assert.equal(scanner.getTokenEnd() - scanner.getTokenStart(), 2);
        }
    }
    assert.equal(scanner.scan(), SyntaxKind.EndOfFile);
});
