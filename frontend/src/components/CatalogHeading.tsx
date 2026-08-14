export function CatalogHeading({ id, label }: { id: string; label: string }) {
  return (
    <div className="catalog-heading">
      <span className="catalog-spine" aria-hidden="true" />
      <div>
        <p className="catalog-label">บัญชีรายการกลาง</p>
        <h2 id={id}>{label}</h2>
      </div>
    </div>
  )
}
