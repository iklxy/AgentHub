// Date: 2026-05-25
// Author: XinYang Li

import type * as React from "react";

declare global {
  namespace JSX {
    type Element = React.JSX.Element;
    interface IntrinsicElements extends React.JSX.IntrinsicElements {}
  }
}

export {};
