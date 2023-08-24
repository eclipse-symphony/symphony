import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";
import { FcSettings } from 'react-icons/fc';
interface PropertyTableProps {
    properties: Record<string, string>;
    refProperties?: Record<string, string>;
}
function SiteList(props: PropertyTableProps) {
    const { properties, refProperties } = props;
    
    //create a new map with elements from properties, pluase a local boolean flag to true
    const combinedProperties: any = {};
    Object.keys(properties).forEach((key: string) => {
        combinedProperties[key] = {
            property: properties[key],
            local: true
        };
    });

    if (refProperties) {
        Object.keys(refProperties).forEach((key: string) => {
            if (!combinedProperties[key]) {
                combinedProperties[key] = {
                    property: refProperties[key],
                    local: false
                };
            }
        });
    }
    return (
        <Table removeWrapper>
            <TableHeader>
                <TableColumn>PROPERTY</TableColumn>
                <TableColumn>VALUE</TableColumn>                
            </TableHeader>
            <TableBody>
                {Object.keys(combinedProperties).slice(0, 5).map((key: string) => (
                    <TableRow key={key} className={combinedProperties[key].local? '': 'remote_row'}>
                        <TableCell>{key}</TableCell>
                        <TableCell>{typeof combinedProperties[key].property === 'string' ? 
                        (combinedProperties[key].property.startsWith('<') ? <div style={{ whiteSpace: 'nowrap' , display: 'inline-flex', gap: '0.5rem', color: 'darkolivegreen'}}>
                            <FcSettings />{combinedProperties[key].property.substring(1, combinedProperties[key].property.length-1)}
                        </div>: combinedProperties[key].property) : '[object]'}</TableCell>
                    </TableRow>
                ))}                         
            </TableBody>
        </Table>
    );
}
export default SiteList;