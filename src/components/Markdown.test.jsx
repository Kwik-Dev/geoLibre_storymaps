import React from 'react';
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import Markdown from './Markdown.jsx';

// A hostile payload must render as inert text — never execute or attach
// event handlers (the XSS boundary is rehype-sanitize).
const HOSTILE_PAYLOADS = [
    '<img src=x onerror=alert(1)>',
    '<script>alert(1)</script>',
    '<a href="javascript:alert(1)">click</a>',
    '<div onclick="alert(1)">hi</div>',
];

describe('Markdown XSS safety', () => {
    it.each(HOSTILE_PAYLOADS)('renders %p as inert (no handler/script survives)', (payload) => {
        const { container } = render(<Markdown text={payload} />);
        const html = container.innerHTML;
        // The XSS boundary: no event handler, script, or javascript: href may
        // survive in the output. rehype-sanitize may drop the hostile element
        // entirely — either way it is inert and never executes.
        expect(html).not.toMatch(/onerror/i);
        expect(html).not.toMatch(/onclick/i);
        expect(html).not.toMatch(/<script/i);
        expect(html).not.toMatch(/javascript:/i);
        // And nothing was actually mounted as a live <script> node.
        expect(container.querySelector('script')).toBeNull();
        expect(container.querySelector('[onerror]')).toBeNull();
        expect(container.querySelector('[onclick]')).toBeNull();
    });

    it('keeps hostile raw HTML inert when embedded in normal markdown text', () => {
        const { container } = render(
            <Markdown text={'Before <img src=x onerror=alert(1)> after'} />,
        );
        const html = container.innerHTML;
        expect(html).not.toMatch(/onerror/i);
        expect(html).not.toMatch(/<script/i);
        // Surrounding prose still renders.
        expect(container.textContent).toContain('Before');
        expect(container.textContent).toContain('after');
    });

    it('renders basic GFM markdown to DOM nodes', () => {
        const { container } = render(<Markdown text={'# Title\n\nSome **bold** and `code`'} />);
        expect(container.querySelector('h1')?.textContent).toBe('Title');
        expect(container.querySelector('strong')?.textContent).toBe('bold');
        expect(container.querySelector('code')?.textContent).toBe('code');
    });

    it('supports GFM (tables, strikethrough, task lists)', () => {
        const { container } = render(
            <Markdown text={'~~strike~~\n\n- [x] done\n- [ ] todo'} />,
        );
        expect(container.querySelector('del')?.textContent).toBe('strike');
        expect(container.querySelectorAll('input[type=checkbox]').length).toBe(2);
    });
});
