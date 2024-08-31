import React, { useState } from 'react';
import { Input, NavbarContent, Button } from '@nextui-org/react';

// Define an interface for the component props
interface FilterProps {
    onSelectFilter: (filter: string) => void; // This function takes a string and returns void
}

const Filter: React.FC<FilterProps> = ({ onSelectFilter }) => {
    const [selectedFilter, setSelectedFilter] = useState(''); // State to keep track of the filter input

    // Function to handle input changes
    const handleFilterChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const filter = event.target.value;
        setSelectedFilter(filter); // Update selected filter state
        onSelectFilter(filter); // Pass filter to parent component
    };

    // Function to handle filter submission
    const handleFilterSubmit = () => {
        onSelectFilter(selectedFilter);
    };

    return (
        <NavbarContent>
            <Input
                isClearable={true}                
                placeholder="Filter"
                value={selectedFilter}
                onChange={handleFilterChange}
            />
            <Button onClick={handleFilterSubmit}>
                Apply
            </Button>
        </NavbarContent>
    );
};

export default Filter;
