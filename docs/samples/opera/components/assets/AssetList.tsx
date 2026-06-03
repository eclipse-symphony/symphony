import { CatalogVersionState } from '../../app/types';
import AssetCard from './AssetCard';
interface AssetListProps {
    catalogversions: CatalogVersionState[];
}
function AssetList({ catalogversions }: AssetListProps) {
    // Create a map of catalogversions by name for easy reference
    const references: Record<string, CatalogVersionState> = {};
    catalogversions.forEach((catalogversion) => {
        references[catalogversion.spec.name] = catalogversion;
    });

    // If you want to merge catalogversions or perform other operations, you can do it directly with the array
    // Assuming mergedCatalogVersions is supposed to be the same as catalogversions in this simplified correction
    const mergedCatalogVersions = [...catalogversions]; // This creates a shallow copy if needed, or directly use catalogversions

    return (
        <div className='sitelist'>
            {mergedCatalogVersions.map((catalogversion) => (
                <AssetCard 
                    key={catalogversion.spec.name} // It's a good practice to provide a unique key for each child in a list
                    catalogversion={catalogversion} 
                    refCatalogVersion={catalogversion.spec.metadata?.['override'] ? references[catalogversion.spec.metadata['override']] : null}
                />
            ))}
        </div>
    );
}
export default AssetList;