# Demo verification

Run `npm test` using Node 22.18+ (verified with Node 24). Tests use Node's built-in
TypeScript stripping and test runner, with no additional test dependencies.

The automated suite covers researcher permission denial, all 36 status pairs,
terminal process locks, editing without changing identity/progress, invalid and
duplicate edits, and deletion including the last item. These are in-memory demo
checks, not verification of production API/database enforcement.

Browser verification, 17 September 2026:

- Administrator can open all six basic-data preview pages. Loading, empty,
  error, and retry states render correctly.
- Researcher has no edit/delete/status/process controls or administrator menu.
  Coordinator can manage research and has no administrator menu.
- Status confirmation displays the selected project, old/new statuses and
  warning. Cancel and Escape retain the old status; confirmation updates it
  and removes earlier statuses from the selector.
- Dialog starts on Cancel; Tab cycles inside the dialog; Escape closes it and
  returns focus to the triggering control.
- Editing the project name preserves its extension status. Terminating a
  project hides status/process mutation controls.
- Cancelling deletion retains the project; confirming removes it and returns
  to the list.
- Desktop details and 390px researcher/coordinator layouts were visually
  inspected. Mobile detail view has no horizontal page overflow; confirmation
  text and actions fit the viewport. Temporary viewport override was reset.
- No browser console errors during the verified flows. Existing reduced-motion
  CSS disables animations/transitions; new dialogs add no animations.

Inline detail editing verification:

- Clicking Edit renders inputs in the existing detail sheet with no dialog.
  Initial focus moves to the project title; Save and Cancel return focus to Edit.
- Editing covers project type, subsidy, both responsible units, start/end dates,
  budget, lead/collaborators, funding type/source, contract, PDF attachment,
  both abstracts, objectives and keywords, in addition to the title.
- Cancel discards title and abstract drafts. Saving updates all text/select
  fields in the detail view and persists them when the project is reopened.
- Duplicate-contract errors keep the draft visible. Researcher views have no
  editable inputs. Status/process controls are disabled during detail editing.
- Automated tests cover attachment replacement, invalid attachments, independent
  draft arrays, and preservation of ID, continuation kind, status and process.
- Desktop and 390px inline-edit layouts were visually inspected; no horizontal
  page overflow or browser console errors occurred in these flows.

Basic-data pages are previews only. All changes and selected files stay in demo
memory and reset on page reload; API and database behavior are unchanged.
