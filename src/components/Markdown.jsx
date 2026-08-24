import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeSanitize from 'rehype-sanitize';

/**
 * Renders user-supplied Markdown (GFM) as real DOM nodes — never
 * dangerouslySetInnerHTML. `rehype-sanitize` is the XSS boundary: any
 * hostile raw HTML embedded in the markdown (e.g. <script>, <img onerror>)
 * is stripped/escaped and renders as inert text.
 *
 *   <Markdown text="# Hello\n\nSome **bold**" />
 */
export default function Markdown({ text }) {
    return (
        <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeSanitize]}
        >
            {text}
        </ReactMarkdown>
    );
}
