/*
* Nanocloud Community, a comprehensive platform to turn any application
* into a cloud solution.
*
* Copyright (C) 2015 Nanocloud Software
*
* This file is part of Nanocloud community.
*
* Nanocloud community is free software; you can redistribute it and/or modify
* it under the terms of the GNU Affero General Public License as
* published by the Free Software Foundation, either version 3 of the
* License, or (at your option) any later version.
*
* Nanocloud community is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
* GNU Affero General Public License for more details.
*
* You should have received a copy of the GNU Affero General Public License
* along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

type GuacamoleXMLConfigs struct {
	XMLName xml.Name             `xml:configs`
	Config  []GuacamoleXMLConfig `xml:"config"`
}

type GuacamoleXMLConfig struct {
	XMLName  xml.Name            `xml:config`
	Name     string              `xml:"name,attr"`
	Protocol string              `xml:"protocol,attr"`
	Params   []GuacamoleXMLParam `xml:"param"`
}

type GuacamoleXMLParam struct {
	ParamName  string `xml:"name,attr"`
	ParamValue string `xml:"value,attr"`
}

type Connection struct {
	Hostname       string `xml:"hostname"`
	Port           string `xml:"port"`
	Username       string `xml:"username"`
	Password       string `xml:"password"`
	RemoteApp      string `xml:"remote-app"`
	ConnectionName string
}

type ApplicationParams struct {
	CollectionName string
	Alias          string
	DisplayName    string
	IconContents   []uint8
	FilePath       string
}

type UserInfo struct {
	Id        string
	Activated bool
	Email     string
	FirstName string
	LastName  string
	IsAdmin   bool
	Sam       string
	Password  string
}

// ========================================================================================================================
// Procedure: CreateConnections
//
// Does:
// - Create all connections in DB for a particular user in order to use all applications
// ========================================================================================================================
func CreateConnections() error {

	type configs GuacamoleXMLConfigs
	var (
		applications    []ApplicationParams
		connections     configs
		executionServer string
	)

	// Seed random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	bashConfigFile, err := os.Create("./scripts/configuration.sh")
	if err != nil {
		log.Printf("Failed to configure bash scripts : %s", err)
		return nil
	}
	err = bashConfigFile.Chmod(0777)
	if err != nil {
		log.Println(err)
	}

	bashConfigFile.Write([]byte("#!/bin/bash\n\n# DO NOT EDIT THIS FILE\n# automatically generated\n\n"))
	bashConfigFile.Write([]byte(fmt.Sprintf("USER=\"%s\"\n", conf.User)))
	bashConfigFile.Write([]byte(fmt.Sprintf("SERVER=\"%s\"\n", conf.Server)))
	bashConfigFile.Write([]byte(fmt.Sprintf("PORT=\"%s\"\n", conf.SSHPort)))
	bashConfigFile.Write([]byte(fmt.Sprintf("PASSWORD=\"%s\"\n", conf.Password)))
	bashConfigFile.Close()
	bashExecScript := "./scripts/exec.sh"
	cmd := exec.Command(bashExecScript, "C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe -Command \"Import-Module RemoteDesktop; Get-RDRemoteApp | ConvertTo-Json -Compress\"")
	cmd.Dir = "."
	response, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to run script exec.sh, error: %s, output: %s\n", err, string(response))
		response = []byte("[]")
	} else if string(response) == "" {
		response = []byte("[]")
	}

	if []byte(response)[0] != []byte("[")[0] {
		response = []byte(fmt.Sprintf("[%s]", string(response)))
	}
	json.Unmarshal(response, &applications)
	for _, application := range applications {
		application.IconContents = []byte(base64.StdEncoding.EncodeToString(application.IconContents))
	}

	//	users, _ := g_Db.GetUsers()
	users := make([]UserInfo, 1)

	users[0] = UserInfo{ //////////////TODO
		Email:    "mail",
		Sam:      "sam",
		Password: "lala",
	}
	for _, user := range users {
		for _, application := range applications {
			if application.Alias == "hapticPowershell" {
				continue
			}

			// Select randomly execution machine from availbale execution machines
			if count := len(conf.ExecutionServers); count > 0 {
				executionServer = conf.ExecutionServers[rand.Intn(count)]
			} else {
				executionServer = conf.Server
			}

			connections.Config = append(connections.Config, GuacamoleXMLConfig{
				Name:     fmt.Sprintf("%s_%s", application.Alias, user.Email),
				Protocol: "rdp",
				Params: []GuacamoleXMLParam{
					GuacamoleXMLParam{
						ParamName:  "hostname",
						ParamValue: executionServer,
					},
					GuacamoleXMLParam{
						ParamName:  "port",
						ParamValue: conf.RDPPort,
					},
					GuacamoleXMLParam{
						ParamName:  "username",
						ParamValue: fmt.Sprintf("%s@%s", user.Sam, conf.WindowsDomain),
					},
					GuacamoleXMLParam{
						ParamName:  "password",
						ParamValue: user.Password,
					},
					GuacamoleXMLParam{
						ParamName:  "remote-app",
						ParamValue: fmt.Sprintf("||%s", application.Alias),
					},
				},
			})
		}
	}

	connections.Config = append(connections.Config, GuacamoleXMLConfig{
		Name:     "hapticDesktop",
		Protocol: "rdp",
		Params: []GuacamoleXMLParam{
			GuacamoleXMLParam{
				ParamName:  "hostname",
				ParamValue: conf.Server,
			},
			GuacamoleXMLParam{
				ParamName:  "port",
				ParamValue: conf.RDPPort,
			},
			GuacamoleXMLParam{
				ParamName:  "username",
				ParamValue: fmt.Sprintf("%s@%s", conf.User, conf.WindowsDomain),
			},
			GuacamoleXMLParam{
				ParamName:  "password",
				ParamValue: conf.Password,
			},
		},
	})
	connections.Config = append(connections.Config, GuacamoleXMLConfig{
		Name:     "hapticPowershell",
		Protocol: "rdp",
		Params: []GuacamoleXMLParam{
			GuacamoleXMLParam{
				ParamName:  "hostname",
				ParamValue: conf.Server,
			},
			GuacamoleXMLParam{
				ParamName:  "port",
				ParamValue: conf.RDPPort,
			},
			GuacamoleXMLParam{
				ParamName:  "username",
				ParamValue: fmt.Sprintf("%s@%s", conf.User, conf.WindowsDomain),
			},
			GuacamoleXMLParam{
				ParamName:  "password",
				ParamValue: conf.Password,
			},
			GuacamoleXMLParam{
				ParamName:  "remote-app",
				ParamValue: "||hapticPowershell",
			},
		},
	})

	output, err := xml.MarshalIndent(connections, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	if err = ioutil.WriteFile(conf.XMLConfigurationFile, output, 0777); err != nil {
		log.Printf("Failed to save connections in %s params : %s", conf.XMLConfigurationFile, err)
		return err
	}

	return nil
}

// ========================================================================================================================
// Procedure: ListApplications
//
// Does:
// - Return list of applications published by Active Directory
// ========================================================================================================================
func ListApplications(reply *PlugRequest) []Connection {
	var (
		guacamoleConfigs GuacamoleXMLConfigs
		connections      []Connection
		bytesRead        []byte
		err              error
	)

	err = CreateConnections()
	if err != nil {
		reply.Status = 500
	}

	if bytesRead, err = ioutil.ReadFile(conf.XMLConfigurationFile); err != nil {
		reply.Status = 500
		log.Printf("Failed to read connections params : %s", err)
	}

	err = xml.Unmarshal(bytesRead, &guacamoleConfigs)
	if err != nil {
		fmt.Printf("error: %v", err)
		reply.Status = 500
		return nil
	}

	for _, config := range guacamoleConfigs.Config {
		var connection Connection

		for _, param := range config.Params {
			switch true {
			case param.ParamName == "hostname":
				connection.Hostname = param.ParamValue
			case param.ParamName == "port":
				connection.Port = param.ParamValue
			case param.ParamName == "username":
				connection.Username = param.ParamValue
			case param.ParamName == "password":
				connection.Password = param.ParamValue
			case param.ParamName == "remote-app":
				connection.RemoteApp = param.ParamValue
			}
		}
		connection.ConnectionName = config.Name

		if connection.RemoteApp == "" || connection.RemoteApp == "||hapticPowershell" {
			continue
		}

		connections = append(connections, connection)
	}

	return connections
}

// ========================================================================================================================
// Procedure: ListApplicationsForSamAccount
//
// Does:
// - Return list of applications available for a particular SAM account
// ========================================================================================================================
func ListApplicationsForSamAccount(sam string, reply *PlugRequest) []Connection {

	var (
		guacamoleConfigs GuacamoleXMLConfigs
		connections      []Connection
		bytesRead        []byte
		err              error
	)

	if bytesRead, err = ioutil.ReadFile(conf.XMLConfigurationFile); err != nil {
		reply.Status = 500
		log.Printf("Failed to read connections params : %s", err)
	}

	err = xml.Unmarshal(bytesRead, &guacamoleConfigs)
	if err != nil {
		fmt.Printf("error: %v", err)
		reply.Status = 500
		return nil
	}

	for _, config := range guacamoleConfigs.Config {
		var connection Connection

		if connection.ConnectionName == "hapticPowershell" {
			continue
		}

		connection.ConnectionName = config.Name
		for _, param := range config.Params {
			switch true {
			case param.ParamName == "hostname":
				connection.Hostname = param.ParamValue
			case param.ParamName == "port":
				connection.Port = param.ParamValue
			case param.ParamName == "username":
				connection.Username = param.ParamValue
			case param.ParamName == "password":
				connection.Password = param.ParamValue
			case param.ParamName == "remote-app":
				connection.RemoteApp = param.ParamValue
			}
		}

		if connection.Username == fmt.Sprintf("%s@%s", sam, conf.WindowsDomain) {
			connections = append(connections, connection)
		}
	}

	return connections
}

// ========================================================================================================================
// Procedure: UnpublishApplication
//
// Does:
// - Unpublish specified applications from ActiveDirectory
// ========================================================================================================================
func UnpublishApp(Alias string) {
	var powershellCmd string

	bashExecScript := "./scripts/exec.sh"
	powershellCmd = fmt.Sprintf(
		"C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe -Command \"Import-Module RemoteDesktop; Remove-RDRemoteApp -Alias %s -CollectionName %s -Force\"",
		Alias,
		"appscollection")
	cmd := exec.Command(bashExecScript, powershellCmd)
	cmd.Dir = (".")
	response, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to run script exec.sh, error: %s, output: %s\n", err, string(response))
	}
}

/*
// ========================================================================================================================
// Procedure: SyncUploadedFile
//
// Does:
// - Upload user files to windows VM
// ========================================================================================================================
func SyncUploadedFile(Filename string) {
	bashCopyScript := filepath.Join(nan.Config().CommonBaseDir, "scripts", "copy.sh")
	cmd := exec.Command(bashCopyScript, Filename)
	cmd.Dir = filepath.Join(nan.Config().CommonBaseDir, "scripts")
	response, err := cmd.Output()
	if err != nil {
		LogError("Failed to run script copy.sh, error: %s, output: %s\n", err, string(response))
	} else {
		Log("SCP upload success for file %s\n", Filename)
	}
}*/
