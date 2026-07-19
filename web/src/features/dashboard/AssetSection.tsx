import type { Asset, ValueType } from "../../api/dashboard";
import { countryForType, currencyForType, formatMoney } from "./format";

interface AssetSectionProps {
  assets: Asset[];
  type: ValueType;
}

export function AssetSection({ assets, type }: AssetSectionProps) {
  const currency = currencyForType(type);
  return (
    <section className="panel asset-panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Portfolio</p>
          <h2>{countryForType(type)}</h2>
        </div>
        <span className="country-badge">{currency}</span>
      </div>
      {assets.length === 0 ? (
        <p className="muted">No {countryForType(type)} assets yet.</p>
      ) : (
        <div className="asset-list">
          {assets.map((asset) => {
            const value = asset.values.find((item) => item.type === type)?.value ?? "0";
            return (
              <div className="asset-row" key={asset.id}>
                <div className="asset-icon">{asset.name.slice(0, 1).toUpperCase()}</div>
                <span>{asset.name}</span>
                <strong>{formatMoney(value, currency)}</strong>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}
