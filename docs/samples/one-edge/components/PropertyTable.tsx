import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";

interface PropertyTableProps {
    properties: Record<string, string>;
}
function SiteList(props: PropertyTableProps) {
    const { properties } = props;
    return (
        <Table removeWrapper>
            <TableHeader>
                <TableColumn>PROPERTY</TableColumn>
                <TableColumn>VALUE</TableColumn>                
            </TableHeader>
            <TableBody>
                {Object.keys(properties).slice(0, 5).map((key: string) => (
                    <TableRow key={key}>
                        <TableCell>{key}</TableCell>
                        <TableCell>{typeof properties[key] === 'string' ? properties[key] : '[object]'}</TableCell>
                    </TableRow>
                ))}                            
            </TableBody>
        </Table>
    );
}
export default SiteList;