// Date: 2026-05-25
// Author: XinYang Li

import { redirect } from "next/navigation";

/**
 * Redirects the root route to the workspace entry page.
 * @returns No visual output because the route redirects immediately.
 */
export default function HomePage(): never {
  redirect("/workspace");
}
