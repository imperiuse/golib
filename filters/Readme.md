# Package Filters

## Intro

The idea of ​​the package is to enable the quick implementation of various custom filters for an HTTP request.

The package already implements all the default methods of the interface filter `Filterer` in the base filter` BaseFilter`.

## Main struct and interfaces

`OrderFilterer` - interface filter order, allows you to link its filter ID list to a filter chain

`Filterer` - the general interface of all filters

`BaseFilter` - the basic structure of the fully implementing interface` Filterer`

## Main idea

Go have not full inheritance, there is only *embedding* and *aggregation*.
 
Based on this, a model with **embedding** the basic structure of the filter and the mechanism for obtaining interface pointers are selected.
myself.


So because I need pointer to child interface for create flexible design and execute child method of parents struct, mechanism of pointer to themself in BaseFilter is used.  
**At the begining you always need to set self pointer by used method!**

    InterfaceCustomFilter.SetSelfPointer(&InterfaceCustomFilter) // ATTENTION! ALWAYS SET THIS! Self pointer! IMPORTANT!

## Usage

Example create New Filter:

    type MyCustomFilter struct {
        filter.BaseFilter
        
        // ... any custom fields
    }

## Examples

@see Examples at `filters_test.go`