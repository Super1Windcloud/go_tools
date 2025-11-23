package  main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	groupID  , artifactId, version, homePage  string

}


func (i item) Title() string       { return i.groupID }
func (i item) Description() string    {
	return fmt.Sprintf("Artifact: %s, Version: %s" +
		"\tHome Page: %s", i.artifactId, i.version , i.homePage ) }
func (i item) FilterValue()     string    {
	if  i  ==( item{}) {
		return  ""
	}
	groupid :=i.groupID
	artifactId :=i.artifactId
	version :=i.version


	return fmt.Sprintf("%s %s %s", groupid, artifactId, version)
}

type model struct {
	list list.Model
	selectedItemInfo string
}

func (m model) Init() tea.Cmd {
	return nil
}
func   refreshModelView(m  model ) {
	if  refreshCount==0 {
		m= model{list: list.New(items , list.NewDefaultDelegate(), 0, 0)}
		m.list.Title = "Maven Dependencies format"

		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String()=="enter" {
			selectedItem  := m.list.SelectedItem().( item) // 获取选中的 item
			_=  selectedItem
			groupid :=selectedItem.groupID
			artifactid :=selectedItem.artifactId
			version :=selectedItem.version
			 templateString  := fmt.Sprintf(`	
			 <dependency>
				 <groupId>%s</groupId>
  	   		  	 <artifactId>%s</artifactId>
  	   	  		 <version>%s</version>
			 </dependency>
			 `, groupid, artifactid, version)
			 _ =templateString
			//  fmt.Println(templateString)
            refreshCount++
         return  m , tea.Quit
		}
		if msg.String()=="down"  || msg.String()=="up" {
			refreshCount   =0 ;
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}
var  refreshCount int
func (m model) View() string {
	view := docStyle.Render(m.list.View())

	return view
}
var items []list.Item

func fancyMavenList(groupIDArr []string, artifactIDArr []string, versionArr []string, homePageArr []string) item {

    for i := 0; i < len(groupIDArr); i++ {
		items = append(items, item{
			groupID: groupIDArr[i],
			artifactId: artifactIDArr[i],
			version: versionArr[i],
			homePage: homePageArr[i],
		})
	}
	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Maven Dependencies format"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return m.list.SelectedItem().(item)
}