// Date: 2026-05-27
// Author: XinYang Li

import { Fragment } from "react";

type InlineSegment =
  | {
      type: "text";
      value: string;
    }
  | {
      type: "strong";
      value: string;
    }
  | {
      type: "code";
      value: string;
    };

type TextContentBlock =
  | {
      type: "paragraph";
      lines: string[];
    }
  | {
      type: "table";
      headers: string[];
      rows: string[][];
    };

/**
 * Splits one markdown line into inline text, bold, and code fragments.
 * Params:
 * - content: the raw line content that may contain `**bold**` or `` `code` `` fragments.
 * Returns:
 * - the ordered inline segment list for rendering.
 */
function splitInlineSegments(content: string): InlineSegment[] {
  const pattern = /(`[^`]+`|\*\*[^*]+\*\*)/g;
  const pieces = content.split(pattern).filter((piece) => piece.length > 0);

  return pieces.map((piece) => {
    if (piece.startsWith("**") && piece.endsWith("**") && piece.length > 4) {
      return {
        type: "strong",
        value: piece.slice(2, -2),
      };
    }

    if (piece.startsWith("`") && piece.endsWith("`") && piece.length > 2) {
      return {
        type: "code",
        value: piece.slice(1, -1),
      };
    }

    return {
      type: "text",
      value: piece,
    };
  });
}

/**
 * Checks whether one line matches a markdown table separator row.
 * Params:
 * - line: the current source line.
 * Returns:
 * - true when the line is a separator row such as `|---|---|`.
 */
function isTableSeparator(line: string): boolean {
  const normalized = line.trim();
  return /^\|?(\s*:?-{3,}:?\s*\|)+\s*:?-{3,}:?\s*\|?$/.test(normalized);
}

/**
 * Parses one markdown table row into trimmed cells.
 * Params:
 * - line: the raw table row that uses `|` separators.
 * Returns:
 * - the trimmed table cell list.
 */
function parseTableRow(line: string): string[] {
  return line
    .trim()
    .replace(/^\|/, "")
    .replace(/\|$/, "")
    .split("|")
    .map((cell) => cell.trim());
}

/**
 * Parses markdown text into paragraph and table blocks.
 * Params:
 * - content: the markdown text block rendered outside fenced code sections.
 * Returns:
 * - the ordered paragraph and table block list.
 */
function parseTextContent(content: string): TextContentBlock[] {
  const lines = content.split("\n");
  const blocks: TextContentBlock[] = [];
  let paragraphBuffer: string[] = [];

  /**
   * Flushes the buffered paragraph lines into the parsed block list.
   * Params:
   * - none: this helper only consumes the outer scoped paragraph buffer.
   * Returns:
   * - nothing. The function mutates the local block collection.
   */
  function flushParagraph(): void {
    if (paragraphBuffer.length === 0) {
      return;
    }

    blocks.push({
      type: "paragraph",
      lines: paragraphBuffer,
    });
    paragraphBuffer = [];
  }

  for (let index = 0; index < lines.length; index += 1) {
    const currentLine = lines[index];
    const nextLine = lines[index + 1];
    const trimmedLine = currentLine.trim();

    if (!trimmedLine) {
      flushParagraph();
      continue;
    }

    if (trimmedLine.includes("|") && nextLine && isTableSeparator(nextLine)) {
      flushParagraph();

      const headers = parseTableRow(currentLine);
      const rows: string[][] = [];
      index += 2;

      while (index < lines.length && lines[index].trim().includes("|")) {
        rows.push(parseTableRow(lines[index]));
        index += 1;
      }

      index -= 1;
      blocks.push({
        type: "table",
        headers,
        rows,
      });
      continue;
    }

    paragraphBuffer.push(currentLine);
  }

  flushParagraph();
  return blocks;
}

/**
 * Renders one reusable markdown content block for chat surfaces.
 * Params:
 * - props.content: the markdown text content rendered inside the current surface.
 * - props.inlineCodeClassName: optional class names applied to inline code segments.
 * - props.paragraphClassName: optional class names applied to paragraph lines.
 * - props.tableClassName: optional class names applied to the outer table wrapper.
 * Returns:
 * - the rendered markdown content.
 */
export function MarkdownContent({
  content,
  inlineCodeClassName,
  paragraphClassName,
  tableClassName,
}: {
  content: string;
  inlineCodeClassName?: string;
  paragraphClassName?: string;
  tableClassName?: string;
}): JSX.Element {
  const blocks = parseTextContent(content);

  return (
    <div className="space-y-3">
      {blocks.map((block, blockIndex) => {
        if (block.type === "table") {
          return (
            <div className={tableClassName ?? "overflow-x-auto rounded-2xl border border-line/80 bg-white/55"} key={`table-${blockIndex}`}>
              <table className="min-w-full border-collapse text-left text-sm leading-7">
                <thead>
                  <tr className="bg-mist/80">
                    {block.headers.map((header, headerIndex) => (
                      <th className="border-b border-line px-4 py-3 font-semibold text-ink" key={`${header}-${headerIndex}`}>
                        {splitInlineSegments(header).map((segment, segmentIndex) =>
                          segment.type === "strong" ? (
                            <strong className="font-semibold" key={`${segment.value}-${segmentIndex}`}>
                              {segment.value}
                            </strong>
                          ) : segment.type === "code" ? (
                            <code className={inlineCodeClassName ?? "rounded-md bg-ink/6 px-1.5 py-0.5 font-mono text-[13px]"} key={`${segment.value}-${segmentIndex}`}>
                              {segment.value}
                            </code>
                          ) : (
                            <Fragment key={`${segment.value}-${segmentIndex}`}>{segment.value}</Fragment>
                          ),
                        )}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {block.rows.map((row, rowIndex) => (
                    <tr className="border-b border-line/70 last:border-b-0" key={`row-${rowIndex}`}>
                      {row.map((cell, cellIndex) => (
                        <td className="px-4 py-3 align-top text-ink/78" key={`${cell}-${cellIndex}`}>
                          {splitInlineSegments(cell).map((segment, segmentIndex) =>
                            segment.type === "strong" ? (
                              <strong className="font-semibold text-ink" key={`${segment.value}-${segmentIndex}`}>
                                {segment.value}
                              </strong>
                            ) : segment.type === "code" ? (
                              <code className={inlineCodeClassName ?? "rounded-md bg-ink/6 px-1.5 py-0.5 font-mono text-[13px] text-ink"} key={`${segment.value}-${segmentIndex}`}>
                                {segment.value}
                              </code>
                            ) : (
                              <Fragment key={`${segment.value}-${segmentIndex}`}>{segment.value}</Fragment>
                            ),
                          )}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          );
        }

        return (
          <div className="space-y-3" key={`paragraph-${blockIndex}`}>
            {block.lines.map((line, lineIndex) => (
              <p className={paragraphClassName ?? "whitespace-pre-wrap text-sm leading-7"} key={`${line}-${lineIndex}`}>
                {splitInlineSegments(line).map((segment, segmentIndex) =>
                  segment.type === "strong" ? (
                    <strong className="font-semibold" key={`${segment.value}-${segmentIndex}`}>
                      {segment.value}
                    </strong>
                  ) : segment.type === "code" ? (
                    <code className={inlineCodeClassName ?? "rounded-md bg-ink/8 px-1.5 py-0.5 font-mono text-[13px] text-current"} key={`${segment.value}-${segmentIndex}`}>
                      {segment.value}
                    </code>
                  ) : (
                    <Fragment key={`${segment.value}-${segmentIndex}`}>{segment.value}</Fragment>
                  ),
                )}
              </p>
            ))}
          </div>
        );
      })}
    </div>
  );
}
